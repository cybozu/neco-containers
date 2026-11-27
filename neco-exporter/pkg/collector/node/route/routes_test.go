package route

import (
	"context"
	"maps"
	"net"
	"testing"
	"time"

	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"

	"github.com/cybozu/neco-containers/neco-exporter/pkg/exporter"
)

const testNode = "node-1"

var t0 = time.Date(2026, 11, 19, 19, 34, 0, 0, time.UTC)

func defaultRoute4(protocol netlink.RouteProtocol, priority int, gws ...string) netlink.Route {
	r := netlink.Route{
		Family:   unix.AF_INET,
		Table:    unix.RT_TABLE_MAIN,
		Protocol: protocol,
		Priority: priority,
	}
	if len(gws) == 1 {
		r.Gw = net.ParseIP(gws[0])
		return r
	}
	for _, gw := range gws {
		r.MultiPath = append(r.MultiPath, &netlink.NexthopInfo{Gw: net.ParseIP(gw)})
	}
	return r
}

func birdDefault(gws ...string) netlink.Route {
	return defaultRoute4(unix.RTPROT_BIRD, 32, gws...)
}

func staticDefault() netlink.Route {
	return defaultRoute4(unix.RTPROT_STATIC, 1024, "10.69.1.193", "10.69.2.1")
}

func update(typ uint16, flags uint16, r netlink.Route) netlink.RouteUpdate {
	return netlink.RouteUpdate{Type: typ, NlFlags: flags, Route: r}
}

func add(r netlink.Route) netlink.RouteUpdate {
	return update(unix.RTM_NEWROUTE, unix.NLM_F_CREATE|unix.NLM_F_EXCL, r)
}

func replace(r netlink.Route) netlink.RouteUpdate {
	return update(unix.RTM_NEWROUTE, unix.NLM_F_CREATE|unix.NLM_F_REPLACE, r)
}

func del(r netlink.Route) netlink.RouteUpdate {
	return update(unix.RTM_DELROUTE, 0, r)
}

// find returns the value of the metric having the name and all the labels.
func find(t *testing.T, ms []*exporter.Metric, name string, labels map[string]string) (float64, bool) {
	t.Helper()
	for _, m := range ms {
		if m.Name != name {
			continue
		}
		matched := true
		for k, v := range labels {
			if m.Labels[k] != v {
				matched = false
				break
			}
		}
		if matched {
			if m.Labels["node"] != testNode {
				t.Errorf("node label = %q, want %q", m.Labels["node"], testNode)
			}
			return m.Value, true
		}
	}
	return 0, false
}

func expect(t *testing.T, ms []*exporter.Metric, name string, labels map[string]string, want float64) {
	t.Helper()
	got, ok := find(t, ms, name, labels)
	if !ok {
		t.Errorf("%s%v not found", name, labels)
		return
	}
	if got != want {
		t.Errorf("%s%v = %v, want %v", name, labels, got, want)
	}
}

func expectMissing(t *testing.T, ms []*exporter.Metric, name string, labels map[string]string) {
	t.Helper()
	if v, ok := find(t, ms, name, labels); ok {
		t.Errorf("%s%v = %v, want no series", name, labels, v)
	}
}

func with(base map[string]string, kv ...string) map[string]string {
	ret := maps.Clone(base)
	for i := 0; i+1 < len(kv); i += 2 {
		ret[kv[i]] = kv[i+1]
	}
	return ret
}

var (
	bird      = map[string]string{"protocol": "bird"}
	birdAdd   = with(bird, "op", opAdd, "default", "true")
	birdRepl  = with(bird, "op", opReplace, "default", "true")
	birdDel   = with(bird, "op", opDelete, "default", "true")
	birdOther = with(bird, "op", opAdd, "default", "false")
	m32       = map[string]string{"metric": "32"}
)

func newTestCollector() *routeCollector {
	c := newRouteCollector()
	c.node = testNode
	return c
}

func TestCollectBeforeFirstDump(t *testing.T) {
	c := newTestCollector()
	c.count(new(add(birdDefault("10.69.1.193"))))
	if _, err := c.Collect(context.Background()); err == nil {
		t.Error("expected an error before the first dump")
	}

	// An empty table is a valid dump.
	c.replaceRoutes(nil, t0)
	if _, err := c.Collect(context.Background()); err != nil {
		t.Errorf("unexpected error after the first dump: %v", err)
	}
}

func TestCountChanges(t *testing.T) {
	c := newTestCollector()
	c.replaceRoutes(nil, t0)
	c.count(new(add(birdDefault("10.69.1.193"))))
	c.count(new(replace(birdDefault("10.69.1.193", "10.69.2.1"))))
	if counted, isDefault := c.count(new(del(birdDefault("10.69.1.193", "10.69.2.1")))); !counted || !isDefault {
		t.Errorf("count() = %v, %v, want true, true", counted, isDefault)
	}

	r := birdDefault("10.69.1.193")
	_, r.Dst, _ = net.ParseCIDR("10.69.0.0/16")
	if counted, isDefault := c.count(new(add(r))); !counted || isDefault {
		t.Errorf("count() = %v, %v, want true, false", counted, isDefault)
	}

	// Routes in other tables, e.g. coil's table 119, and IPv6 routes are ignored.
	other := birdDefault("10.69.1.193")
	other.Table = 119
	v6 := birdDefault("fe80::1")
	v6.Family = unix.AF_INET6
	for _, u := range []netlink.RouteUpdate{add(other), del(other), add(v6)} {
		if counted, _ := c.count(&u); counted {
			t.Errorf("count(%v) = true, want false", u.Route)
		}
	}

	ms := c.metrics()
	expect(t, ms, "changes_total", birdAdd, 1)
	expect(t, ms, "changes_total", birdRepl, 1)
	expect(t, ms, "changes_total", birdDel, 1)
	expect(t, ms, "changes_total", birdOther, 1)
}

func expectReplace(t *testing.T, c *routeCollector, routes []netlink.Route, now time.Time, wantHad, wantHas bool, wantPrev time.Time) {
	t.Helper()
	had, has, prev := c.replaceRoutes(routes, now)
	if had != wantHad || has != wantHas || !prev.Equal(wantPrev) {
		t.Errorf("replaceRoutes() = %v, %v, %v, want %v, %v, %v", had, has, prev, wantHad, wantHas, wantPrev)
	}
}

func TestDefaultDeletedAndAdded(t *testing.T) {
	c := newTestCollector()
	expectReplace(t, c, []netlink.Route{birdDefault("10.69.1.193", "10.69.2.1")}, t0, true, true, time.Time{})

	ms := c.metrics()
	expect(t, ms, "default_present", nil, 1)
	expect(t, ms, "default_nexthops", m32, 2)

	expectReplace(t, c, nil, t0.Add(time.Minute), true, false, t0)

	ms = c.metrics()
	expect(t, ms, "default_present", nil, 0)
	expectMissing(t, ms, "default_nexthops", m32)

	expectReplace(t, c, []netlink.Route{birdDefault("10.69.1.193")}, t0.Add(2*time.Minute), false, true, t0.Add(time.Minute))

	ms = c.metrics()
	expect(t, ms, "default_present", nil, 1)
	expect(t, ms, "default_nexthops", m32, 1)
}

func TestFirstDumpWithoutDefaultRoute(t *testing.T) {
	c := newTestCollector()
	// The node is assumed to have had a default route, so that the loss is logged at the start.
	expectReplace(t, c, nil, t0, true, false, time.Time{})
}

func TestStaticFallback(t *testing.T) {
	c := newTestCollector()
	c.replaceRoutes([]netlink.Route{birdDefault("10.69.1.193", "10.69.2.1"), staticDefault()}, t0)
	expectReplace(t, c, []netlink.Route{staticDefault()}, t0.Add(time.Second), true, true, t0)

	ms := c.metrics()
	// The node still has a default route.
	expect(t, ms, "default_present", nil, 1)
	// BIRD's route (metric 32) is missing, which default_nexthops shows.
	expect(t, ms, "default_nexthops", map[string]string{"metric": "1024"}, 2)
	expectMissing(t, ms, "default_nexthops", m32)
}

func TestDefaultNexthops(t *testing.T) {
	specific := birdDefault("10.69.1.193")
	_, specific.Dst, _ = net.ParseCIDR("10.69.0.0/16")
	zero := staticDefault()
	_, zero.Dst, _ = net.ParseCIDR("0.0.0.0/0")
	// Added with `ip route append`.
	appended := defaultRoute4(unix.RTPROT_STATIC, 1024, "10.69.3.1")
	// One of the devices of BIRD's route is down.
	bird := birdDefault("10.69.1.193", "10.69.2.1")
	bird.MultiPath[1].Flags = unix.RTNH_F_DEAD | unix.RTNH_F_LINKDOWN
	// The carrier of the device of a single-path route is lost.
	linkdown := defaultRoute4(unix.RTPROT_STATIC, 2048, "10.69.4.1")
	linkdown.Flags = unix.RTNH_F_LINKDOWN

	c := newTestCollector()
	c.replaceRoutes([]netlink.Route{specific, zero, appended, bird, linkdown}, t0)

	ms := c.metrics()
	expect(t, ms, "default_present", nil, 1)
	expect(t, ms, "default_nexthops", map[string]string{"metric": "1024"}, 3)
	expect(t, ms, "default_nexthops", m32, 1)
	// The series of a route without usable next hops remains.
	expect(t, ms, "default_nexthops", map[string]string{"metric": "2048"}, 0)
}

func TestFormatNexthop(t *testing.T) {
	tests := []struct {
		nh   netlink.NexthopInfo
		want string
	}{
		{netlink.NexthopInfo{Gw: net.ParseIP("10.69.1.193"), LinkIndex: 2}, "via 10.69.1.193 ifindex 2"},
		{netlink.NexthopInfo{LinkIndex: 1}, "ifindex 1"},
		{netlink.NexthopInfo{Gw: net.ParseIP("10.69.2.1"), LinkIndex: 3, Flags: unix.RTNH_F_DEAD | unix.RTNH_F_LINKDOWN}, "via 10.69.2.1 ifindex 3 dead linkdown"},
	}
	for _, tt := range tests {
		if got := formatNexthop(&tt.nh); got != tt.want {
			t.Errorf("formatNexthop() = %q, want %q", got, tt.want)
		}
	}
}

func TestMapRoutes(t *testing.T) {
	zero := staticDefault()
	_, zero.Dst, _ = net.ParseCIDR("0.0.0.0/0")
	// Added with `ip route append`, without Dst.
	appended := defaultRoute4(unix.RTPROT_STATIC, 1024, "10.69.3.1")
	specific := birdDefault("10.69.1.193")
	_, specific.Dst, _ = net.ParseCIDR("10.69.0.0/16")

	m := mapRoutes([]netlink.Route{zero, appended, specific, birdDefault("10.69.1.193")})

	if len(m) != 3 {
		t.Errorf("len(mapRoutes()) = %d, want 3", len(m))
	}
	if rs := m[keyOf(&appended)]; len(rs) != 2 {
		t.Errorf("routes sharing the key of the static default route = %d, want 2", len(rs))
	}
	if !keyOf(&zero).isDefault() || keyOf(&specific).isDefault() {
		t.Error("isDefault() is wrong")
	}
}
