package route

import (
	"net/netip"
	"strconv"
	"strings"
	"time"

	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"

	"github.com/cybozu/neco-containers/neco-exporter/pkg/exporter"
)

// Label values of op in changes_total.
const (
	opAdd     = "add"
	opReplace = "replace"
	opDelete  = "delete"
)

// changeLabels are the labels of changes_total except node.
type changeLabels struct {
	protocol  string
	op        string
	isDefault bool
}

func isIPv4MainTable(r *netlink.Route) bool {
	return r.Family == unix.AF_INET && r.Table == unix.RT_TABLE_MAIN
}

func isDefault(r *netlink.Route) bool {
	if r.Dst == nil {
		return true
	}
	ones, _ := r.Dst.Mask.Size()
	return ones == 0
}

// routeKey identifies an IPv4 route in the main table, the same way the kernel does.
type routeKey struct {
	dst      netip.Prefix
	tos      int
	priority int
}

func keyOf(r *netlink.Route) routeKey {
	dst := netip.PrefixFrom(netip.IPv4Unspecified(), 0)
	if r.Dst != nil {
		addr, _ := netip.AddrFromSlice(r.Dst.IP)
		ones, _ := r.Dst.Mask.Size()
		dst = netip.PrefixFrom(addr.Unmap(), ones)
	}
	return routeKey{dst: dst, tos: r.Tos, priority: r.Priority}
}

func (k routeKey) isDefault() bool {
	return k.dst.Bits() == 0
}

// mapRoutes maps the routes by their keys.
// Routes added with `ip route append` share a key.
func mapRoutes(routes []netlink.Route) map[routeKey][]netlink.Route {
	ret := make(map[routeKey][]netlink.Route, len(routes))
	for _, r := range routes {
		k := keyOf(&r)
		ret[k] = append(ret[k], r)
	}
	return ret
}

func hasDefault(routes map[routeKey][]netlink.Route) bool {
	for k := range routes {
		if k.isDefault() {
			return true
		}
	}
	return false
}

// nexthopsOf returns the next hops of the route.
// The next hop of a single-path route is in the route itself, including its flags.
func nexthopsOf(r *netlink.Route) []*netlink.NexthopInfo {
	if len(r.MultiPath) > 0 {
		return r.MultiPath
	}
	return []*netlink.NexthopInfo{{LinkIndex: r.LinkIndex, Gw: r.Gw, Flags: r.Flags}}
}

// isUsable returns false if the kernel marked the next hop dead or linkdown, as `ip route` shows.
// The kernel keeps such next hops in multipath routes, without notifications, while their devices are down.
func isUsable(nh *netlink.NexthopInfo) bool {
	return nh.Flags&(unix.RTNH_F_DEAD|unix.RTNH_F_LINKDOWN) == 0
}

// formatNexthop formats the next hop for logs, like `ip route` does.
func formatNexthop(nh *netlink.NexthopInfo) string {
	var b strings.Builder
	if nh.Gw != nil {
		b.WriteString("via " + nh.Gw.String() + " ")
	}
	b.WriteString("ifindex " + strconv.Itoa(nh.LinkIndex))
	if nh.Flags&unix.RTNH_F_DEAD != 0 {
		b.WriteString(" dead")
	}
	if nh.Flags&unix.RTNH_F_LINKDOWN != 0 {
		b.WriteString(" linkdown")
	}
	return b.String()
}

func opOf(u *netlink.RouteUpdate) string {
	if u.Type == unix.RTM_DELROUTE {
		return opDelete
	}
	if u.NlFlags&unix.NLM_F_REPLACE != 0 {
		return opReplace
	}
	return opAdd
}

// count counts a route notification in changes_total.
// It returns whether the notification is about the main table, and whether it is about a default route.
func (c *routeCollector) count(u *netlink.RouteUpdate) (counted, isDefaultRoute bool) {
	r := &u.Route
	if !isIPv4MainTable(r) {
		return false, false
	}
	if u.Type != unix.RTM_NEWROUTE && u.Type != unix.RTM_DELROUTE {
		return false, false
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	def := isDefault(r)
	c.changes[changeLabels{protocol: r.Protocol.String(), op: opOf(u), isDefault: def}]++
	return true, def
}

// replaceRoutes replaces the routes with a dump of the main table taken at now.
// It returns whether the node had and has default routes, and the time of the previous dump.
// Before the first dump, the node is assumed to have a default route,
// so that the first dump reports the loss if the node has none.
func (c *routeCollector) replaceRoutes(routes []netlink.Route, now time.Time) (had, has bool, prev time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()

	m := mapRoutes(routes)
	had = c.dumpedAt.IsZero() || hasDefault(c.routes)
	has = hasDefault(m)
	prev = c.dumpedAt
	c.routes = m
	c.dumpedAt = now
	return had, has, prev
}

func boolValue(b bool) float64 {
	if b {
		return 1
	}
	return 0
}

// metrics returns the metrics from the last dump and the counts of the notifications.
func (c *routeCollector) metrics() []*exporter.Metric {
	c.mu.Lock()
	defer c.mu.Unlock()

	ret := []*exporter.Metric{{
		Name:   "default_present",
		Labels: map[string]string{"node": c.node},
		Value:  boolValue(hasDefault(c.routes)),
	}}

	// Routes differing only in TOS, or added with `ip route append`, share the labels, so their next hops are summed.
	nexthops := make(map[int]int)
	for k, routes := range c.routes {
		if !k.isDefault() {
			continue
		}
		nexthops[k.priority] += 0
		for i := range routes {
			for _, nh := range nexthopsOf(&routes[i]) {
				if isUsable(nh) {
					nexthops[k.priority]++
				}
			}
		}
	}
	for priority, n := range nexthops {
		ret = append(ret, &exporter.Metric{
			Name:   "default_nexthops",
			Labels: map[string]string{"node": c.node, "metric": strconv.Itoa(priority)},
			Value:  float64(n),
		})
	}

	for l, v := range c.changes {
		ret = append(ret, &exporter.Metric{
			Name: "changes_total",
			Labels: map[string]string{
				"node":     c.node,
				"protocol": l.protocol,
				"op":       l.op,
				"default":  strconv.FormatBool(l.isDefault),
			},
			Value: v,
		})
	}
	return ret
}
