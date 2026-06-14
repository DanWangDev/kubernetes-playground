// Package kubectl provides a Go wrapper around the kubectl CLI.
// Exercises use this package rather than calling os/exec directly,
// which gives consistent error handling and a cleaner API.
//
// Design note: we intentionally shell out to kubectl rather than using
// client-go. The goal of this playground is kubectl fluency — students
// should learn the real commands they'll use every day in production.
package kubectl

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// DefaultTimeout is the default wait timeout for operations like WaitForPodReady.
const DefaultTimeout = 120 * time.Second

// run executes a kubectl command and returns stdout or an error with stderr.
func run(args ...string) (string, error) {
	cmd := exec.Command("kubectl", args...)

	// Pass through kubeconfig env if set
	if kc := os.Getenv("KUBECONFIG"); kc != "" {
		cmd.Env = append(os.Environ(), "KUBECONFIG="+kc)
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("kubectl %s: %w\n%s", strings.Join(args, " "), err, stderr.String())
	}
	return stdout.String(), nil
}

// Apply applies a YAML manifest file to the cluster.
func Apply(manifestPath string) error {
	_, err := run("apply", "-f", manifestPath)
	return err
}

// ApplyString applies inline YAML to the cluster. Useful for tests
// and exercises that construct manifests programmatically.
func ApplyString(yaml string) error {
	cmd := exec.Command("kubectl", "apply", "-f", "-")
	cmd.Stdin = strings.NewReader(yaml)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if kc := os.Getenv("KUBECONFIG"); kc != "" {
		cmd.Env = append(os.Environ(), "KUBECONFIG="+kc)
	}
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("kubectl apply -f -: %w\n%s", err, stderr.String())
	}
	return nil
}

// Delete removes resources defined in a YAML manifest file.
func Delete(manifestPath string) error {
	_, err := run("delete", "-f", manifestPath, "--ignore-not-found")
	return err
}

// DeleteResource removes a specific resource by kind, name, and optionally namespace.
func DeleteResource(kind, name, namespace string) error {
	args := []string{"delete", kind, name, "--ignore-not-found"}
	if namespace != "" {
		args = append(args, "-n", namespace)
	}
	_, err := run(args...)
	return err
}

// Get runs a kubectl get command and returns the output as a string.
// Example: Get("pods", "-n", "playground")
func Get(resource string, flags ...string) (string, error) {
	args := append([]string{"get", resource}, flags...)
	return run(args...)
}

// GetJSON returns the JSON representation of a resource.
// Example: GetJSON("pod", "nginx", "-n", "playground")
func GetJSON(resource string, flags ...string) (map[string]interface{}, error) {
	args := append([]string{"get", resource}, flags...)
	args = append(args, "-o", "json")
	out, err := run(args...)
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		return nil, fmt.Errorf("parsing kubectl JSON output: %w", err)
	}
	return result, nil
}

// Describe returns the describe output for a resource.
func Describe(resource, name string, flags ...string) (string, error) {
	args := append([]string{"describe", resource, name}, flags...)
	return run(args...)
}

// Logs retrieves logs from a pod container.
func Logs(name string, container string, flags ...string) (string, error) {
	args := []string{"logs", name}
	if container != "" {
		args = append(args, "-c", container)
	}
	args = append(args, flags...)
	return run(args...)
}

// Exec runs a command inside a pod container and returns stdout.
// Extra flags (e.g., "-n", "playground") are passed through to kubectl.
func Exec(name string, command []string, container string, flags ...string) (string, error) {
	args := []string{"exec", name}
	if container != "" {
		args = append(args, "-c", container)
	}
	args = append(args, flags...)
	args = append(args, "--")
	args = append(args, command...)
	return run(args...)
}

// Wait blocks until a condition is met on a resource.
func Wait(resource, name, condition, timeout string) error {
	args := []string{"wait", resource, name}
	if condition != "" {
		args = append(args, "--for", condition)
	}
	if timeout != "" {
		args = append(args, "--timeout", timeout)
	}
	_, err := run(args...)
	return err
}

// WaitForPodReady blocks until a pod reports Ready status.
func WaitForPodReady(name, namespace string, timeout time.Duration) error {
	args := []string{"wait", "pod", name}
	if namespace != "" {
		args = append(args, "-n", namespace)
	}
	args = append(args, "--for", "condition=Ready", "--timeout", timeout.String())
	_, err := run(args...)
	return err
}

// WaitForDeployReady blocks until all deployment replicas are available.
func WaitForDeployReady(name, namespace string, timeout time.Duration) error {
	args := []string{"wait", "deployment", name}
	if namespace != "" {
		args = append(args, "-n", namespace)
	}
	args = append(args, "--for", "condition=Available", "--timeout", timeout.String())
	_, err := run(args...)
	return err
}

// RolloutStatus waits for a rollout to complete.
func RolloutStatus(resource, name string) error {
	_, err := run("rollout", "status", resource, name)
	return err
}

// CreateNamespace creates a new namespace.
func CreateNamespace(name string) error {
	_, err := run("create", "namespace", name, "--dry-run=client", "-o", "yaml")
	if err != nil {
		return err
	}
	return ApplyString(fmt.Sprintf(`apiVersion: v1
kind: Namespace
metadata:
  name: %s
`, name))
}

// DeleteNamespace deletes a namespace and all resources within it.
func DeleteNamespace(name string) error {
	_, err := run("delete", "namespace", name, "--ignore-not-found", "--timeout=60s")
	return err
}

// NamespaceExists checks if a namespace exists.
func NamespaceExists(name string) bool {
	_, err := run("get", "namespace", name)
	return err == nil
}

// Scale sets the replica count for a deployment or statefulset.
func Scale(resource, name string, replicas int, namespace string) error {
	args := []string{"scale", resource, name, fmt.Sprintf("--replicas=%d", replicas)}
	if namespace != "" {
		args = append(args, "-n", namespace)
	}
	_, err := run(args...)
	return err
}

// ClusterInfo returns kubectl cluster-info output.
func ClusterInfo() (string, error) {
	return run("cluster-info")
}

// GetNodes returns the list of nodes.
func GetNodes() (string, error) {
	return run("get", "nodes")
}

// GetPodPhase returns the phase of a pod (Running, Pending, Succeeded, Failed).
func GetPodPhase(name, namespace string) (string, error) {
	args := []string{"get", "pod", name}
	if namespace != "" {
		args = append(args, "-n", namespace)
	}
	args = append(args, "-o", "jsonpath={.status.phase}")
	return run(args...)
}

// GetDeployReplicas returns ready/total replicas for a deployment.
func GetDeployReplicas(name, namespace string) (ready, total string, err error) {
	args := []string{"get", "deployment", name}
	if namespace != "" {
		args = append(args, "-n", namespace)
	}
	args = append(args, "-o", "jsonpath={.status.readyReplicas}/{.status.replicas}")
	out, err := run(args...)
	if err != nil {
		return "", "", err
	}
	parts := strings.SplitN(strings.TrimSpace(out), "/", 2)
	if len(parts) == 2 {
		return parts[0], parts[1], nil
	}
	return strings.TrimSpace(out), "", nil
}
