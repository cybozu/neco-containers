// Package route reports the IPv4 default routes in the node's main routing table and changes of the IPv4 routes.
// It watches route notifications from the kernel, so that it detects changes shorter than the collection interval.
package route

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"

	"github.com/cybozu/neco-containers/neco-exporter/pkg/constants"
	"github.com/cybozu/neco-containers/neco-exporter/pkg/exporter"
)

const (
	nodeNameEnv = "NODE_NAME"

	// receiveBufferSize is the requested receive buffer of the notification socket.
	// The kernel caps it at net.core.rmem_max.
	receiveBufferSize = 4 << 20

	// updateQueueSize is the size of the channel of route notifications from the netlink socket.
	updateQueueSize = 4096

	// defaultDumpInterval is the minimum interval of the dumps triggered by notifications of default routes.
	// The dumps are rare, as default routes rarely change, so they run soon to log the transitions in time.
	defaultDumpInterval = 100 * time.Millisecond

	// dumpInterval is the minimum interval of the dumps triggered by notifications of the other routes.
	// It bounds the cost of the dumps while routes keep changing, e.g. when BIRD reconverges.
	dumpInterval = time.Second

	// resyncInterval is the maximum interval of the dumps.
	// Linux does not notify some removals, e.g. routes removed when their device goes down.
	resyncInterval = 30 * time.Second

	// restartDelay is the delay before subscribing again after the subscription stops.
	restartDelay = time.Second
)

func subscribe(ch chan<- netlink.RouteUpdate, done <-chan struct{}, onError func(error)) error {
	return netlink.RouteSubscribeWithOptions(ch, done, netlink.RouteSubscribeOptions{
		ErrorCallback:     onError,
		ReceiveBufferSize: receiveBufferSize,
	})
}

// list dumps the IPv4 routes in the main table.
func list() ([]netlink.Route, error) {
	return netlink.RouteListFiltered(netlink.FAMILY_V4, &netlink.Route{Table: unix.RT_TABLE_MAIN}, netlink.RT_FILTER_TABLE)
}

// routeCollector keeps the last dump of the IPv4 main table and counts the route notifications.
// The routes are always taken from dumps, not from notifications.
// Notifications are not versioned, so ones received during a dump may be older than the dump.
// They only trigger dumps.
type routeCollector struct {
	node   string
	logger *slog.Logger

	mu sync.Mutex
	// routes is the last dump, taken at dumpedAt, which is zero until the first dump.
	routes   map[routeKey][]netlink.Route
	dumpedAt time.Time
	changes  map[changeLabels]float64
}

var _ exporter.Collector = &routeCollector{}

// NewCollector returns a node-scope collector of the IPv4 default routes.
// The Pod must run in the host network namespace to watch the node's routing table.
func NewCollector() exporter.Collector {
	return newRouteCollector()
}

func newRouteCollector() *routeCollector {
	return &routeCollector{changes: make(map[changeLabels]float64)}
}

func (c *routeCollector) Scope() string {
	return constants.ScopeNode
}

func (c *routeCollector) MetricsPrefix() string {
	return "route"
}

func (c *routeCollector) IsLeaderMetrics() bool {
	return false
}

func (c *routeCollector) Setup(ctx context.Context) error {
	node := os.Getenv(nodeNameEnv)
	if node == "" {
		return fmt.Errorf("%s environment variable is not set", nodeNameEnv)
	}
	c.node = node
	// The collectors are created before main sets the default logger, so get it here.
	if c.logger == nil {
		c.logger = slog.Default()
	}

	// Fail early if the routing table cannot be read at all,
	// and report the routes from the first collection, which runs before the watcher dumps them.
	err := c.dump(ctx, time.Time{})
	if err != nil && !errors.Is(err, netlink.ErrDumpInterrupted) {
		return fmt.Errorf("failed to list routes: %w", err)
	}

	go c.watch(ctx)
	return nil
}

func (c *routeCollector) Collect(_ context.Context) ([]*exporter.Metric, error) {
	c.mu.Lock()
	synced := !c.dumpedAt.IsZero()
	c.mu.Unlock()
	if !synced {
		return nil, errors.New("routes have not been listed yet")
	}
	return c.metrics(), nil
}

// watch keeps the subscription until ctx is cancelled.
func (c *routeCollector) watch(ctx context.Context) {
	for {
		err := c.watchOnce(ctx)
		if ctx.Err() != nil {
			return
		}
		// Notifications are lost until the dump after subscribing again.
		c.logger.WarnContext(ctx, "route subscription stopped; subscribing again", slog.Any("error", err))

		select {
		case <-ctx.Done():
			return
		case <-time.After(restartDelay):
		}
	}
}

// watchOnce subscribes to route notifications and dumps the main table when they arrive, until the subscription stops.
// It returns the error that stopped it, e.g. ENOBUFS when the receive buffer overflowed.
func (c *routeCollector) watchOnce(ctx context.Context) error {
	ch := make(chan netlink.RouteUpdate, updateQueueSize)
	done := make(chan struct{})
	defer close(done)

	var mu sync.Mutex
	var lastErr error
	onError := func(err error) {
		mu.Lock()
		defer mu.Unlock()
		lastErr = err
	}

	if err := subscribe(ch, done, onError); err != nil {
		return err
	}

	// What happened since the last dump, to decide when to dump next.
	var (
		lastTry    time.Time
		failed     bool
		changed    bool
		notifiedAt time.Time // the first notification of a default route, or zero
	)
	receive := func(u *netlink.RouteUpdate) {
		counted, isDefault := c.count(u)
		changed = changed || counted
		if isDefault {
			if notifiedAt.IsZero() {
				notifiedAt = time.Now()
			}
			c.logNotification(ctx, u)
		}
	}
	dump := func() {
		lastTry = time.Now()
		if err := c.dump(ctx, notifiedAt); err != nil {
			// Keep what happened for the next dump.
			failed = true
			return
		}
		failed, changed, notifiedAt = false, false, time.Time{}
	}
	untilNext := func() time.Duration {
		interval := resyncInterval
		switch {
		case failed:
			// Avoid retrying a failing dump too often.
			interval = dumpInterval
		case !notifiedAt.IsZero():
			interval = defaultDumpInterval
		case changed:
			interval = dumpInterval
		}
		return time.Until(lastTry.Add(interval))
	}

	// Subscribe first, then dump, so that no change is missed between them.
	// The changes during the subscription loss were not notified, so the dump does not report when they happened.
	dump()
	timer := time.NewTimer(untilNext())
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case u, ok := <-ch:
			if !ok {
				mu.Lock()
				defer mu.Unlock()
				if lastErr == nil {
					return errors.New("subscription closed")
				}
				return lastErr
			}
			receive(&u)
		case <-timer.C:
			// The notifications already received are reflected in the dump.
			for range len(ch) {
				u, ok := <-ch
				if !ok {
					break
				}
				receive(&u)
			}
			dump()
		}
		timer.Reset(untilNext())
	}
}

// dump replaces the routes with a dump and logs when the node loses all of its default routes and when one comes back.
// notifiedAt is the time of the first notification of a default route since the previous dump, or zero.
// The logs tell when the node had no default route, even if they reach the log storage after the routes come back.
func (c *routeCollector) dump(ctx context.Context, notifiedAt time.Time) error {
	routes, err := list()
	if err != nil {
		// A dump is interrupted when the table changes during it, e.g. while BIRD reconverges.
		// The change triggers another dump.
		if !errors.Is(err, netlink.ErrDumpInterrupted) {
			c.logger.WarnContext(ctx, "failed to list routes", slog.Any("error", err))
		}
		return err
	}

	had, has, prev := c.replaceRoutes(routes, time.Now())
	if had == has {
		return nil
	}

	// The change happened after notBefore: the first notification, or the previous dump if the change was not notified.
	detectedBy, notBefore := "notification", notifiedAt
	if notifiedAt.IsZero() {
		detectedBy, notBefore = "dump", prev
	}
	attrs := []slog.Attr{slog.String("detected_by", detectedBy)}
	if !notBefore.IsZero() {
		attrs = append(attrs, slog.Time("not_before", notBefore))
	}
	if has {
		c.logger.LogAttrs(ctx, slog.LevelInfo, "default route is restored", attrs...)
	} else {
		c.logger.LogAttrs(ctx, slog.LevelWarn, "default route is missing", attrs...)
	}
	return nil
}

// logNotification logs a notification of a default route.
// Unlike the logs of dump, it shows every change, including ones merged by a dump and ones leaving other default routes.
func (c *routeCollector) logNotification(ctx context.Context, u *netlink.RouteUpdate) {
	r := &u.Route
	var nexthops []string
	for _, nh := range nexthopsOf(r) {
		nexthops = append(nexthops, formatNexthop(nh))
	}
	c.logger.InfoContext(ctx, "default route changed",
		slog.String("op", opOf(u)),
		slog.Int("metric", r.Priority),
		slog.String("protocol", r.Protocol.String()),
		slog.Any("nexthops", nexthops))
}
