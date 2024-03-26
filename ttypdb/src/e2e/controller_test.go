package e2e

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	policyv1 "k8s.io/api/policy/v1"
)

const configuredPollingIntervalSeconds = 5
const testIntervalSeconds = configuredPollingIntervalSeconds + 2
const testInterval = time.Second * time.Duration(testIntervalSeconds)

func TestE2E(t *testing.T) {
	// this test is intended to be run from a terminal by hand.
	// If you want to run from CI environment, PTY should be used.
	tty, err := os.Open("/dev/tty")
	if err != nil {
		t.Fatal("cannot open /dev/tty", err)
	}
	defer tty.Close()

	// prepare

	_, stderr, err := kubectl("apply", "-f", "../../role.yaml")
	if err != nil {
		t.Fatal("failed to apply role.yaml", err, string(stderr))
	}
	_, stderr, err = kubectl("apply", "-f", "../../e2e.yaml")
	if err != nil {
		t.Fatal("failed to apply e2e.yaml", err, string(stderr))
	}

	// PDBs should not exist in initial state

	time.Sleep(testInterval)

	stdout, stderr, err := kubectl("get", "pdb", "--ignore-not-found", "-ojson")
	if err != nil {
		t.Fatal("failed to get pdb list", err, string(stderr))
	}
	if len(stdout) != 0 {
		t.Error("unexpected pdb exists", string(stdout))
	}

	// a PDB should be created for teststs-0 because it is selected by `-l` option of the controller

	go func() {
		_, stderr, err := kubectlWithReaderStdin(tty, "exec", "teststs-0", "-it", "--", "sleep", fmt.Sprintf("%d", testIntervalSeconds))
		if err != nil {
			t.Error("failed to login to pod", err, string(stderr))
		}
	}()

	time.Sleep(testInterval)

	fmt.Println("PDB should be created for teststs-0")
	stdout, stderr, err = kubectl("get", "pdb", "--ignore-not-found", "-ojson")
	if err != nil {
		t.Fatal("failed to get pdb list", err, string(stderr))
	}
	if len(stdout) == 0 {
		t.Error("expected pdb does not exist")
	}
	pdbList := policyv1.PodDisruptionBudgetList{}
	err = json.Unmarshal(stdout, &pdbList)
	if err != nil {
		t.Fatal("failed to unmarshal json", err)
	}
	if len(pdbList.Items) != 1 {
		t.Error("expected pdb does not exist", pdbList.Items)
	}
	if pdbList.Items[0].Name != "teststs-0" {
		t.Error("expected pdb does not exist", pdbList.Items)
	}
	if pdbList.Items[0].Spec.Selector.MatchLabels["statefulset.kubernetes.io/pod-name"] != "teststs-0" {
		t.Error("expected pdb does not exist", pdbList.Items)
	}

	// the PDB should be deleted after the logout from the Pod

	time.Sleep(testInterval)

	fmt.Println("PDB should be deleted")
	stdout, stderr, err = kubectl("get", "pdb", "--ignore-not-found", "-ojson")
	if err != nil {
		t.Fatal("failed to get pdb list", err, string(stderr))
	}
	if len(stdout) != 0 {
		t.Error("unexpected pdb exists", string(stdout))
	}

	// a PDB should not be created for teststs2-0 because it is not selected by `-l` option of the controller

	go func() {
		_, stderr, err := kubectlWithReaderStdin(tty, "exec", "teststs2-0", "-it", "--", "sleep", fmt.Sprintf("%d", testIntervalSeconds))
		if err != nil {
			t.Error("failed to login to pod", err, string(stderr))
		}
	}()

	time.Sleep(testInterval)

	stdout, stderr, err = kubectl("get", "pdb", "--ignore-not-found", "-ojson")
	if err != nil {
		t.Fatal("failed to get pdb list", err, string(stderr))
	}
	if len(stdout) != 0 {
		t.Error("unexpected pdb exists", string(stdout))
	}

	// cleanup

	_, stderr, err = kubectl("delete", "-f", "../../e2e.yaml")
	if err != nil {
		t.Fatal(err, string(stderr))
	}
}
