// Package validate provides assertion and polling helpers for exercises.
// Each function checks an expected condition and returns an error with
// a descriptive message if the condition is not met.
package validate

import (
	"fmt"
	"strings"
	"time"

	"github.com/danwa/kubernetes-playground/pkg/kubectl"
)

// WaitForCondition polls a check function until it returns true or times out.
// Interval is the time between checks.
func WaitForCondition(check func() (bool, error), timeout, interval time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		ok, err := check()
		if err != nil {
			return err
		}
		if ok {
			return nil
		}
		time.Sleep(interval)
	}
	return fmt.Errorf("condition not met after %v", timeout)
}

// AssertPodPhase checks that a pod is in the expected phase.
func AssertPodPhase(name, namespace, expectedPhase string) error {
	phase, err := kubectl.GetPodPhase(name, namespace)
	if err != nil {
		return fmt.Errorf("getting pod phase for %s: %w", name, err)
	}
	if strings.TrimSpace(phase) != expectedPhase {
		return fmt.Errorf("pod %s: expected phase %q, got %q", name, expectedPhase, strings.TrimSpace(phase))
	}
	return nil
}

// AssertDeployReady checks that a deployment is fully available.
func AssertDeployReady(name, namespace string) error {
	ready, total, err := kubectl.GetDeployReplicas(name, namespace)
	if err != nil {
		return fmt.Errorf("getting deploy status for %s: %w", name, err)
	}
	if ready != total || ready == "0" {
		return fmt.Errorf("deployment %s: expected all replicas ready, got %s/%s", name, ready, total)
	}
	return nil
}

// AssertResourceExists checks that a resource exists in the cluster.
func AssertResourceExists(kind, name, namespace string) error {
	args := []string{name}
	if namespace != "" {
		args = append(args, "-n", namespace)
	}
	_, err := kubectl.Get(kind, args...)
	if err != nil {
		return fmt.Errorf("%s %s not found in namespace %s: %w", kind, name, namespace, err)
	}
	return nil
}
