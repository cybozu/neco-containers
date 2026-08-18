package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	k8sresource "k8s.io/apimachinery/pkg/api/resource"
)

const (
	kindContext = "kind-cpu-lending-agent"
	nodeName    = "cpu-lending-agent-control-plane"
	resource    = "cpu-lending.cybozu.io/preemptible-millicpu"
)

func kubectlBin() string {
	if v := os.Getenv("KUBECTL"); v != "" {
		return v
	}
	return "kubectl"
}

func kubectl(input []byte, args ...string) ([]byte, error) {
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	cmd := exec.Command(kubectlBin(), append([]string{"--context", kindContext}, args...)...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	if input != nil {
		cmd.Stdin = bytes.NewReader(input)
	}
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("kubectl %v failed with %s: stderr=%s", args, err, stderr)
	}
	return stdout.Bytes(), nil
}

// nodeExec runs a shell command inside the kind node container.
func nodeExec(command string) (string, error) {
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	cmd := exec.Command("docker", "exec", nodeName, "bash", "-c", command)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("docker exec %q failed with %s: stderr=%s", command, err, stderr)
	}
	return strings.TrimSpace(stdout.String()), nil
}

// parseCPUList parses a kernel cpuset list ("0-2,5") into a set.
func parseCPUList(s string) (map[int]bool, error) {
	set := map[int]bool{}
	s = strings.TrimSpace(s)
	if s == "" {
		return set, nil
	}
	for _, part := range strings.Split(s, ",") {
		lo, hi, ranged := strings.Cut(part, "-")
		start, err := strconv.Atoi(lo)
		if err != nil {
			return nil, fmt.Errorf("invalid cpuset %q: %w", s, err)
		}
		end := start
		if ranged {
			end, err = strconv.Atoi(hi)
			if err != nil {
				return nil, fmt.Errorf("invalid cpuset %q: %w", s, err)
			}
		}
		for i := start; i <= end; i++ {
			set[i] = true
		}
	}
	return set, nil
}

// lentCPUs returns the exclusive CPUs of the mysqld container from the
// kubelet checkpoint on the node.
func lentCPUs() (map[int]bool, error) {
	out, err := nodeExec("cat /var/lib/kubelet/cpu_manager_state")
	if err != nil {
		return nil, err
	}
	var state struct {
		Entries map[string]map[string]string `json:"entries"`
	}
	if err := json.Unmarshal([]byte(out), &state); err != nil {
		return nil, err
	}
	for _, containers := range state.Entries {
		if cpus, ok := containers["mysqld"]; ok {
			return parseCPUList(cpus)
		}
	}
	return nil, fmt.Errorf("mysqld not found in cpu_manager_state: %s", out)
}

// podCPUSet returns the actual cgroup cpuset of the first container of the
// named pod.
func podCPUSet(pod string) (map[int]bool, error) {
	out, err := kubectl(nil, "get", "pod", pod, "-o", "jsonpath={.metadata.uid}")
	if err != nil {
		return nil, err
	}
	uid := strings.ReplaceAll(string(out), "-", "_")
	cpus, err := nodeExec(fmt.Sprintf(
		`find /sys/fs/cgroup -path "*pod%s*" -name cpuset.cpus | grep -v effective | xargs cat | grep -v "^$" | head -1`, uid))
	if err != nil {
		return nil, err
	}
	return parseCPUList(cpus)
}

func intersects(a, b map[int]bool) bool {
	for cpu := range a {
		if b[cpu] {
			return true
		}
	}
	return false
}

func containsAll(super, sub map[int]bool) bool {
	for cpu := range sub {
		if !super[cpu] {
			return false
		}
	}
	return true
}

// advertised returns the capacity and allocatable values of the extended
// resource on the node, in milli-CPUs. Values are compared numerically
// because quantity serialization may canonicalize (2000 vs "2k").
func advertised() (capacity, allocatable int64, err error) {
	esc := strings.ReplaceAll(resource, ".", `\.`)
	out, err := kubectl(nil, "get", "node", nodeName, "-o",
		fmt.Sprintf("jsonpath={.status.capacity.%s},{.status.allocatable.%s}", esc, esc))
	if err != nil {
		return 0, 0, err
	}
	capStr, allocStr, _ := strings.Cut(string(out), ",")
	capQ, err := k8sresource.ParseQuantity(capStr)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid capacity %q: %w", capStr, err)
	}
	allocQ, err := k8sresource.ParseQuantity(allocStr)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid allocatable %q: %w", allocStr, err)
	}
	return capQ.Value(), allocQ.Value(), nil
}
