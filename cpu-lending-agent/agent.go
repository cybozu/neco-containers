package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/containerd/nri/pkg/api"
	"github.com/containerd/nri/pkg/stub"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	listerscorev1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/record"
)

// Config holds the agent configuration.
type Config struct {
	NodeName       string
	StatePath      string
	LenderSelector labels.Selector
	RoleLabel      string
	LendRoleValue  string
	// LenderContainer is the container whose exclusive CPUs are lent
	// ("mysqld"); other pinned containers in the pod keep their exclusivity.
	LenderContainer string
	// ResourceName is the extended resource that advertises the lending
	// capacity and whose request marks a pod as a borrower.
	ResourceName string
	// BaselineCacheFile persists the last known shared pool (empty to
	// disable). It lets a freshly restarted agent reclaim lent CPUs even
	// while the kubelet checkpoint is unreadable.
	BaselineCacheFile string
	// StatusAnnotation is the node annotation key for the human-readable
	// lending summary (empty to disable).
	StatusAnnotation string
	WholeCoresOnly   bool
	ResyncPeriod     time.Duration
	RetryInterval    time.Duration
}

// trackedContainer is a container on this node, tracked via NRI events.
type trackedContainer struct {
	id           string
	podNamespace string
	podName      string
	podUID       string
	bestEffort   bool   // derived from the pod cgroup parent
	current      CPUSet // the cpuset we believe is applied
}

// Agent implements the NRI hooks and the level-triggered reconcile loop.
//
// Inputs: the kubelet CPU manager checkpoint (exclusive assignments and
// shared pool), the pod informer (role labels of lenders and resource
// requests of borrowers on this node), and NRI events (container lifecycle).
// Output: cpuset rewrites of borrower containers and the capacity
// advertisement on the Node status, nothing else. The agent never touches
// lender (mysqld) containers or the kubelet state.
type Agent struct {
	cfg        Config
	topo       *Topology
	pods       listerscorev1.PodLister
	stub       stub.Stub
	advertiser *NodeAdvertiser // nil when capacity advertisement is disabled
	recorder   record.EventRecorder
	nodeRef    *corev1.ObjectReference
	logger     *slog.Logger

	kick chan struct{}

	mu         sync.Mutex
	containers map[string]*trackedContainer
	baseline   CPUSet // kubelet shared pool (defaultCpuSet)
	lendable   CPUSet // CPUs currently lent to borrowers
	inert      bool   // true when the kubelet policy is not static
}

// NewAgent creates an Agent. SetStub must be called before Run.
func NewAgent(cfg Config, topo *Topology, pods listerscorev1.PodLister, logger *slog.Logger) *Agent {
	return &Agent{
		cfg:        cfg,
		topo:       topo,
		pods:       pods,
		logger:     logger,
		kick:       make(chan struct{}, 1),
		containers: map[string]*trackedContainer{},
		baseline:   CPUSet{},
		lendable:   CPUSet{},
	}
}

// SetStub injects the NRI stub used for unsolicited container updates.
func (a *Agent) SetStub(s stub.Stub) { a.stub = s }

// SetAdvertiser enables extended-resource capacity advertisement.
func (a *Agent) SetAdvertiser(adv *NodeAdvertiser) { a.advertiser = adv }

// SetRecorder enables node events for lending transitions.
func (a *Agent) SetRecorder(r record.EventRecorder) {
	a.recorder = r
	// The UID is deliberately the node NAME, following the kubelet
	// convention: `kubectl describe node` searches events with
	// involvedObject.uid=<node name> and would not show them otherwise.
	a.nodeRef = &corev1.ObjectReference{Kind: "Node", Name: a.cfg.NodeName, UID: types.UID(a.cfg.NodeName)}
}

// Kick schedules a reconcile. It never blocks.
func (a *Agent) Kick() {
	select {
	case a.kick <- struct{}{}:
	default:
	}
}

// track records or refreshes a container seen via an NRI event.
func (a *Agent) track(pod *api.PodSandbox, ctr *api.Container, current CPUSet) *trackedContainer {
	t := &trackedContainer{
		id:           ctr.GetId(),
		podNamespace: pod.GetNamespace(),
		podName:      pod.GetName(),
		podUID:       pod.GetUid(),
		bestEffort:   strings.Contains(pod.GetLinux().GetCgroupParent(), "besteffort"),
		current:      current,
	}
	a.containers[t.id] = t
	return t
}

// isBorrower reports whether the tracked container belongs to a borrower
// pod: BestEffort QoS and at least one container requesting the preemptible
// CPU resource. The request is the single opt-in signal; requesting it
// declares that the workload tolerates losing the CPUs (and its node) at any
// time. Returns false when the pod is not yet in the informer cache; the
// level-triggered reconcile converges such containers shortly after.
func (a *Agent) isBorrower(t *trackedContainer) bool {
	if !t.bestEffort {
		return false
	}
	pod, err := a.pods.Pods(t.podNamespace).Get(t.podName)
	if err != nil {
		return false
	}
	// Stable pod names (StatefulSets) can be reused by a new incarnation
	// while the old container still exists; never evaluate a container
	// against another incarnation's spec.
	if string(pod.UID) != t.podUID {
		return false
	}
	return RequestsResource(pod, a.cfg.ResourceName)
}

// RequestsResource reports whether any container of the pod requests the
// named resource.
func RequestsResource(pod *corev1.Pod, resourceName string) bool {
	for _, c := range pod.Spec.Containers {
		if q, ok := c.Resources.Requests[corev1.ResourceName(resourceName)]; ok && !q.IsZero() {
			return true
		}
	}
	return false
}

func currentCPUs(ctr *api.Container) CPUSet {
	set, err := ParseCPUSet(ctr.GetLinux().GetResources().GetCpu().GetCpus())
	if err != nil {
		return CPUSet{}
	}
	return set
}

// Synchronize rebuilds the container inventory when (re)connecting to
// containerd and returns updates that converge borrowers to the desired
// state. This is the recovery path after agent or containerd restarts.
func (a *Agent) Synchronize(_ context.Context, pods []*api.PodSandbox, containers []*api.Container) ([]*api.ContainerUpdate, error) {
	podByID := map[string]*api.PodSandbox{}
	for _, pod := range pods {
		podByID[pod.GetId()] = pod
	}

	if err := a.refreshLending(); err != nil {
		// Keep fail-closed defaults (lend nothing) rather than failing the
		// synchronization; the reconcile loop will retry.
		a.logger.Error("synchronize: failed to refresh lending state", "error", err)
	}

	a.mu.Lock()
	defer a.mu.Unlock()
	a.containers = map[string]*trackedContainer{}
	var updates []*api.ContainerUpdate
	for _, ctr := range containers {
		pod := podByID[ctr.GetPodSandboxId()]
		if pod == nil {
			continue
		}
		t := a.track(pod, ctr, currentCPUs(ctr))
		if !a.isBorrower(t) {
			continue
		}
		desired := DesiredBorrowerCPUSet(a.baseline, a.lendable)
		if len(desired) != 0 && !t.current.Equal(desired) {
			u := &api.ContainerUpdate{ContainerId: t.id}
			u.SetLinuxCPUSetCPUs(desired.String())
			updates = append(updates, u)
			// t.current is intentionally NOT updated here: Synchronize gives
			// no per-container feedback on whether the update was applied.
			// The first reconcile re-verifies through the tracked path (an
			// identical duplicate update is idempotent and harmless).
			a.logger.Info("synchronize: converging borrower", "pod", t.podName, "container", ctr.GetName(), "cpus", desired.String())
		}
	}
	return updates, nil
}

// CreateContainer gives borrower containers their initial cpuset:
// the kubelet shared pool plus the currently lendable CPUs.
func (a *Agent) CreateContainer(_ context.Context, pod *api.PodSandbox, ctr *api.Container) (*api.ContainerAdjustment, []*api.ContainerUpdate, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	t := a.track(pod, ctr, currentCPUs(ctr))
	if !a.isBorrower(t) {
		return nil, nil, nil
	}

	desired := DesiredBorrowerCPUSet(a.baseline, a.lendable)
	if len(desired) == 0 {
		// No baseline known yet; leave the container untouched. The kubelet
		// sends its own update shortly after start, which we intercept.
		return nil, nil, nil
	}
	t.current = desired
	adjust := &api.ContainerAdjustment{}
	adjust.SetLinuxCPUSetCPUs(desired.String())
	a.logger.Info("create: assigned borrower cpuset", "pod", pod.GetName(), "container", ctr.GetName(), "cpus", desired.String())
	return adjust, nil, nil
}

// UpdateContainer intercepts runtime-requested updates (typically the kubelet
// CPU manager writing the shared pool) and re-injects the lent CPUs for
// borrower containers. The requested cpuset is taken as the new baseline for
// this container: when lending is off this rewrite is the identity.
func (a *Agent) UpdateContainer(_ context.Context, pod *api.PodSandbox, ctr *api.Container, res *api.LinuxResources) ([]*api.ContainerUpdate, error) {
	// Updates that do not carry a cpuset (memory-only changes, quota-only
	// changes) must pass through untouched: an absent field means "keep the
	// current cpuset", not "empty cpuset".
	if res.GetCpu().GetCpus() == "" {
		return nil, nil
	}
	requested, err := ParseCPUSet(res.GetCpu().GetCpus())
	if err != nil {
		a.logger.Error("update: unparsable requested cpuset", "pod", pod.GetName(), "container", ctr.GetName(), "error", err)
		return nil, nil
	}

	a.mu.Lock()
	defer a.mu.Unlock()
	t, ok := a.containers[ctr.GetId()]
	if !ok {
		t = a.track(pod, ctr, requested)
	}
	if !a.isBorrower(t) {
		t.current = requested
		return nil, nil
	}

	desired := DesiredBorrowerCPUSet(requested, a.lendable)
	// This hook gets no confirmation that the rewrite is applied, so the
	// applied value is unknown: recording `desired` would wrongly skip a
	// retry when the write failed, and recording `requested` would wrongly
	// skip the next reclaim when it succeeded (the cgroup holds the lend,
	// but reconcile would believe it is already at the baseline). Mark the
	// state unknown; the next reconcile always re-verifies through the
	// confirmed unsolicited-update path (one idempotent write).
	t.current = nil
	u := &api.ContainerUpdate{ContainerId: ctr.GetId()}
	u.SetLinuxCPUSetCPUs(desired.String())
	a.logger.Info("update: rewrote borrower cpuset",
		"pod", pod.GetName(), "container", ctr.GetName(),
		"requested", requested.String(), "cpus", desired.String())
	return []*api.ContainerUpdate{u}, nil
}

// RemoveContainer prunes tracking state.
func (a *Agent) RemoveContainer(_ context.Context, _ *api.PodSandbox, ctr *api.Container) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	delete(a.containers, ctr.GetId())
	return nil
}

// refreshLending recomputes the baseline and lendable sets from the kubelet
// checkpoint and the lender pods on this node. On error it fails closed:
// the lendable set is cleared (while keeping the last known baseline) so
// that the ongoing reconcile reclaims lent CPUs instead of keeping a stale
// lending decision alive during e.g. a kubelet state outage.
func (a *Agent) refreshLending() error {
	st, err := LoadPinnedState(a.cfg.StatePath)
	if err != nil {
		a.failClosed()
		return fmt.Errorf("failed to load kubelet CPU manager state: %w", err)
	}

	var lenders []Lender
	pods, err := a.pods.List(a.cfg.LenderSelector)
	if err != nil {
		a.failClosed()
		return fmt.Errorf("failed to list lender pods: %w", err)
	}
	for _, pod := range pods {
		role, ok := pod.Labels[a.cfg.RoleLabel]
		lenders = append(lenders, Lender{
			UID: string(pod.UID),
			// Fail-closed: a missing role label (failover in progress)
			// means "do not lend".
			Lending: ok && role == a.cfg.LendRoleValue,
		})
	}

	lendable, err := ComputeLendable(st, lenders, a.cfg.LenderContainer, a.cfg.WholeCoresOnly, a.topo)
	if err != nil {
		a.failClosed()
		return err
	}

	if st.Static && a.cfg.BaselineCacheFile != "" {
		if err := SaveBaselineCache(a.cfg.BaselineCacheFile, st.SharedPool); err != nil {
			a.logger.Error("failed to save baseline cache", "error", err)
		}
	}

	a.mu.Lock()
	oldLendable := a.lendable
	wasInert := a.inert
	if st.Static {
		// The baseline is only meaningful under the static policy; under
		// "none" the defaultCpuSet is empty and the kubelet does not manage
		// cpusets at all, so the last static baseline is kept untouched.
		a.baseline = st.SharedPool
	}
	a.lendable = lendable
	a.inert = !st.Static
	if a.inert {
		metricInert.Set(1)
		a.logger.Error("kubelet CPU manager policy is not static; lending disabled (inert mode)")
	} else {
		metricInert.Set(0)
	}
	metricLendableCPUs.Set(float64(len(lendable)))
	a.mu.Unlock()

	// Emit node events on transitions only, so that `kubectl describe node`
	// keeps a short history of when lending changed and why the numbers
	// moved (events complement the point-in-time status annotation).
	if a.recorder != nil {
		if !oldLendable.Equal(lendable) {
			cpus := lendable.String()
			if cpus == "" {
				cpus = "none"
			}
			a.recorder.Eventf(a.nodeRef, corev1.EventTypeNormal, "CPULendingChanged",
				"lending capacity changed from %d to %d milli-CPUs (lendable CPUs: %s)",
				len(oldLendable)*milliCPUsPerCPU, len(lendable)*milliCPUsPerCPU, cpus)
		}
		if !st.Static && !wasInert {
			a.recorder.Eventf(a.nodeRef, corev1.EventTypeWarning, "CPULendingInert",
				"kubelet CPU manager policy is not static; lending disabled")
		}
	}
	return nil
}

// failClosed clears the lendable set while keeping the last known baseline,
// so that convergence reclaims lent CPUs. When no baseline is known (agent
// restarted while the kubelet checkpoint is unreadable), it falls back to
// the persisted baseline cache; without this, stale lends of a previous
// agent incarnation would survive until the checkpoint becomes readable.
func (a *Agent) failClosed() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.lendable = CPUSet{}
	metricLendableCPUs.Set(0)
	if len(a.baseline) != 0 || a.cfg.BaselineCacheFile == "" {
		return
	}
	cached, err := LoadBaselineCache(a.cfg.BaselineCacheFile)
	if err != nil {
		a.logger.Error("failed to load baseline cache", "error", err)
		return
	}
	a.baseline = cached
}

// reconcile converges every borrower container to the desired cpuset, then
// advertises the lending capacity (reality first, ledger second).
//
// A refreshLending error does not abort the run: refreshLending has already
// failed closed (lendable cleared), and proceeding lets this reconcile
// reclaim lent CPUs and shrink the advertised capacity right away. The error
// is still returned afterwards so that the caller retries.
func (a *Agent) reconcile(ctx context.Context) error {
	metricReconciles.Inc()
	refreshErr := a.refreshLending()

	a.mu.Lock()
	desired := DesiredBorrowerCPUSet(a.baseline, a.lendable)
	var updates []*api.ContainerUpdate
	var targets []*trackedContainer
	borrowers := 0
	for _, t := range a.containers {
		if !a.isBorrower(t) {
			continue
		}
		borrowers++
		// Inert (non-static kubelet policy): write nothing. Under "none"
		// every pod may use every CPU, so forcing borrowers back to an old
		// static-era shared pool would be wrong; the agent's contract is to
		// go silent and alert.
		if a.inert {
			continue
		}
		// An empty desired set means the baseline is not known yet (e.g.
		// startup with an unreadable kubelet state); leave the kubelet
		// values in place rather than writing an empty cpuset. A nil
		// current means the applied value is unknown (a hook rewrite went
		// unconfirmed): always re-verify with an idempotent update.
		if len(desired) == 0 || (t.current != nil && t.current.Equal(desired)) {
			continue
		}
		u := &api.ContainerUpdate{ContainerId: t.id}
		u.SetLinuxCPUSetCPUs(desired.String())
		updates = append(updates, u)
		targets = append(targets, t)
	}
	lendableStr := a.lendable.String()
	a.mu.Unlock()
	metricBorrowers.Set(float64(borrowers))

	if len(updates) == 0 {
		metricUnconverged.Set(0)
		return errors.Join(a.advertise(ctx), refreshErr)
	}
	a.logger.Info("reconcile: updating borrowers", "count", len(updates), "lendable", lendableStr, "cpus", desired.String())

	// stub.UpdateContainers is a synchronous RPC to containerd; do not hold
	// the mutex across it.
	failed, err := a.stub.UpdateContainers(updates)
	failedIDs := map[string]struct{}{}
	for _, f := range failed {
		failedIDs[f.GetContainerId()] = struct{}{}
	}

	a.mu.Lock()
	converged := true
	for _, t := range targets {
		if _, ok := failedIDs[t.id]; ok || err != nil {
			// Leave t.current unchanged so the next reconcile retries.
			converged = false
			continue
		}
		t.current = desired
		metricUpdates.Inc()
	}
	a.mu.Unlock()

	var updateErr error
	switch {
	case err != nil:
		metricUnconverged.Set(1)
		updateErr = fmt.Errorf("failed to update %d borrower containers: %w", len(updates), err)
	case !converged:
		metricUnconverged.Set(1)
		updateErr = fmt.Errorf("failed to update %d of %d borrower containers", len(failedIDs), len(updates))
	default:
		metricUnconverged.Set(0)
	}

	// Advertise even when some container updates failed: a.lendable is the
	// value the create/update hooks already act on, and on the reclaim path
	// the capacity must shrink immediately so that the scheduler stops
	// placing new borrowers here.
	return errors.Join(updateErr, a.advertise(ctx), refreshErr)
}

// milliCPUsPerCPU is the advertised quantity per lent logical CPU. The
// extended resource is denominated in milli-CPUs (1000 = one lent CPU),
// following the same convention as Koordinator's kubernetes.io/batch-cpu.
// This is a fixed unit, not a knob: the unit is a cluster-wide contract
// between the agent and every borrower manifest.
const milliCPUsPerCPU = 1000

// advertise publishes the current lending capacity as a node extended
// resource. It runs after cpuset convergence so that the ledger never gets
// ahead of reality: on reclaim the capacity shrinks only after the CPUs are
// actually taken back.
func (a *Agent) advertise(ctx context.Context) error {
	if a.advertiser == nil {
		return nil
	}
	a.mu.Lock()
	milli := len(a.lendable) * milliCPUsPerCPU
	a.mu.Unlock()
	if err := a.advertiser.Ensure(ctx, milli); err != nil {
		return fmt.Errorf("failed to advertise lending capacity: %w", err)
	}
	metricAdvertised.Set(float64(milli))
	return a.annotate(ctx, milli)
}

// annotate maintains a human-readable lending summary on the node so that
// `kubectl describe node` shows the headroom with the subtraction already
// done, e.g.:
//
//	capacity_milli=2000 allocated_milli=1000 free_milli=1000
//
// A negative free_milli is the over-allocated state after a reclaim and
// resolves as borrowers finish. Lender details are deliberately omitted
// (nodes can host many MySQL pods); look them up via the pod role labels.
func (a *Agent) annotate(ctx context.Context, capacityMilli int) error {
	if a.cfg.StatusAnnotation == "" {
		return nil
	}
	pods, err := a.pods.List(labels.Everything())
	if err != nil {
		return fmt.Errorf("failed to list pods for the status annotation: %w", err)
	}
	allocated := int64(0)
	for _, pod := range pods {
		if pod.Status.Phase == corev1.PodSucceeded || pod.Status.Phase == corev1.PodFailed {
			continue
		}
		for _, c := range pod.Spec.Containers {
			if q, ok := c.Resources.Requests[corev1.ResourceName(a.cfg.ResourceName)]; ok {
				allocated += q.Value()
			}
		}
	}
	value := fmt.Sprintf("capacity_milli=%d allocated_milli=%d free_milli=%d",
		capacityMilli, allocated, int64(capacityMilli)-allocated)
	if err := a.advertiser.EnsureAnnotation(ctx, a.cfg.StatusAnnotation, value); err != nil {
		return fmt.Errorf("failed to update the status annotation: %w", err)
	}
	return nil
}

// Run is the reconcile loop: level-triggered on kicks (label changes, pod
// events) with a periodic resync as a safety net. Errors are retried with a
// fixed interval.
func (a *Agent) Run(ctx context.Context) error {
	ticker := time.NewTicker(a.cfg.ResyncPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-a.kick:
		case <-ticker.C:
		}
		if err := a.reconcile(ctx); err != nil {
			metricReconcileErrors.Inc()
			a.logger.Error("reconcile failed", "error", err)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(a.cfg.RetryInterval):
				a.Kick()
			}
		}
	}
}
