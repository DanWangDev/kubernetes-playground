// Module 01: First Pods — Exercise Runner
//
// Run:
//
//	go run ./exercises/01-first-pods/             (automatic)
//	go run ./exercises/01-first-pods/ --step      (interactive step-by-step)
//
// Prerequisites:
//
//	make cluster-create    (or: go run ./cmd/cluster-create/)
//	make cluster-status    (verify cluster is ready)
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/danwa/kubernetes-playground/pkg/kubectl"
	"github.com/danwa/kubernetes-playground/pkg/logger"
	"github.com/danwa/kubernetes-playground/pkg/prompt"
)

var (
	log         = logger.New("01-first-pods")
	manifestDir = filepath.Join("exercises", "01-first-pods", "manifests")
	ns          = "playground"
)

func main() {
	stepMode := flag.Bool("step", false, "Run interactively step by step")
	flag.Parse()
	if *stepMode {
		prompt.EnableStepMode()
	}

	log.Section("Module 01: First Pods")
	log.Info("Domain: Pod lifecycle, kubectl essentials, labels, multi-container pods")
	log.Info("Duration: ~5 minutes")
	fmt.Println()

	// ── Step 0: Verify prerequisites ──────────────────────────────
	log.Section("Step 0: Cluster Check")
	log.Step("Verifying the cluster is reachable...")
	info, err := kubectl.ClusterInfo()
	if err != nil {
		log.Error("Cannot reach cluster. Run: make cluster-create")
		os.Exit(1)
	}
	log.Output(info)
	log.Success("Cluster is reachable!")
	prompt.StepPause()

	// ── Step 1: Create namespace ──────────────────────────────────
	log.Section("Step 1: Create a Namespace")
	log.Concept(
		"Namespaces partition your cluster into virtual sub-clusters.\n" +
			"All exercises in this playground will use the 'playground' namespace\n" +
			"to keep everything organized and easy to clean up.\n" +
			"\n" +
			"Think of namespaces as folders in a filesystem — they group related\n" +
			"resources and provide scope for names and access controls.",
	)

	// Check if namespace already exists
	if kubectl.NamespaceExists(ns) {
		log.Warn("Namespace %q already exists — skipping creation.", ns)
	} else {
		log.Step("Creating namespace %q...", ns)
		manifest := filepath.Join(manifestDir, "01-namespace.yaml")
		if err := kubectl.Apply(manifest); err != nil {
			log.Error("Failed to create namespace: %v", err)
			os.Exit(1)
		}
		log.Success("Namespace %q created!", ns)
	}

	fmt.Println()
	log.Command("kubectl get namespaces")
	out, _ := kubectl.Get("namespaces")
	log.Output(out)
	prompt.StepPause()

	// ── Step 2: Launch a simple Pod ───────────────────────────────
	log.Section("Step 2: Launch Your First Pod")
	log.Concept(
		"A Pod is the smallest deployable unit in Kubernetes. It wraps one\n" +
			"or more containers that share network and storage.\n" +
			"\n" +
			"This manifest creates a single-container Pod running nginx:alpine.\n" +
			"The Pod will go through: Pending -> Running.",
	)

	log.Step("Applying simple Pod manifest...")
	manifest := filepath.Join(manifestDir, "02-simple-pod.yaml")
	if err := kubectl.Apply(manifest); err != nil {
		log.Error("Failed to create pod: %v", err)
		os.Exit(1)
	}
	log.Success("Pod 'nginx' created!")

	log.Step("Waiting for Pod to be ready...")
	if err := kubectl.WaitForPodReady("nginx", ns, kubectl.DefaultTimeout); err != nil {
		log.Error("Pod did not become ready: %v", err)
		os.Exit(1)
	}
	log.Success("Pod 'nginx' is Running!")

	fmt.Println()
	log.Command("kubectl get pods -n playground")
	out, _ = kubectl.Get("pods", "-n", ns)
	log.Output(out)

	// Show pod details
	log.Command("kubectl describe pod nginx -n playground")
	desc, _ := kubectl.Describe("pod", "nginx", "-n", ns)
	// Print first 30 lines of describe output
	lines := splitLines(desc, 30)
	log.Output(lines)
	prompt.StepPause()

	// ── Step 3: Exec into the Pod ─────────────────────────────────
	log.Section("Step 3: Exec Into the Running Pod")
	log.Concept(
		"kubectl exec lets you run commands inside a container, just like SSH.\n" +
			"This is invaluable for debugging: checking configs, testing connectivity,\n" +
			"and inspecting the filesystem.",
	)

	log.Step("Running 'nginx -v' inside the pod...")
	out, err = kubectl.Exec("nginx", []string{"nginx", "-v"}, "", "-n", ns)
	if err != nil {
		// nginx -v writes to stderr, which kubectl captures
		log.Warn("(nginx -v writes to stderr — expected)")
	}
	log.Command("kubectl exec nginx -n playground -- nginx -v")
	log.Output(out)

	log.Step("Checking the nginx welcome page from inside the pod...")
	out, err = kubectl.Exec("nginx", []string{"wget", "-q", "-O", "-", "http://localhost"}, "", "-n", ns)
	if err != nil {
		log.Warn("(wget may not be installed in nginx:alpine)")
	} else {
		log.Output(firstLine(out))
		log.Success("nginx is serving content on port 80!")
	}
	prompt.StepPause()

	// ── Step 4: Labels and selectors ──────────────────────────────
	log.Section("Step 4: Labels and Selectors")
	log.Concept(
		"Labels are key-value pairs attached to Kubernetes objects. They are the\n" +
			"primary mechanism for grouping, selecting, and connecting resources.\n" +
			"\n" +
			"Services use label selectors to find their target Pods.\n" +
			"Deployments use label selectors to track their managed Pods.\n" +
			"You use label selectors with kubectl -l to filter output.",
	)

	log.Step("Creating pod with rich labels...")
	manifest = filepath.Join(manifestDir, "03-pod-with-labels.yaml")
	if err := kubectl.Apply(manifest); err != nil {
		log.Error("Failed to create labeled pod: %v", err)
		os.Exit(1)
	}
	log.Success("Pod 'nginx-labeled' created!")

	// Show all pods with labels
	log.Command("kubectl get pods -n playground --show-labels")
	out, _ = kubectl.Get("pods", "-n", ns, "--show-labels")
	log.Output(out)

	// Query by label
	fmt.Println()
	log.Command("kubectl get pods -n playground -l app=web")
	out, _ = kubectl.Get("pods", "-n", ns, "-l", "app=web")
	log.Output(out)

	log.Info("Labels let you query, group, and connect resources without")
	log.Info("hardcoding names or IPs. This loose coupling is key to Kubernetes design.")
	prompt.StepPause()

	// ── Step 5: Multi-container Pod ───────────────────────────────
	log.Section("Step 5: Multi-Container Pod (Sidecar Pattern)")
	log.Concept(
		"A Pod can run multiple containers that share fate and resources:\n" +
			"  - Same network namespace: reach each other at localhost\n" +
			"  - Same IPC namespace: shared memory, semaphores\n" +
			"  - Shared volumes: mounted filesystems\n" +
			"\n" +
			"Common patterns: sidecar (log shipper, proxy), ambassador (broker proxy),\n" +
			"and adapter (output normalizer).",
	)

	log.Step("Creating multi-container pod (nginx + debug sidecar)...")
	manifest = filepath.Join(manifestDir, "04-multi-container-pod.yaml")
	if err := kubectl.Apply(manifest); err != nil {
		log.Error("Failed to create multi-container pod: %v", err)
		os.Exit(1)
	}

	log.Step("Waiting for multi-container pod...")
	if err := kubectl.WaitForPodReady("multi-container", ns, kubectl.DefaultTimeout); err != nil {
		log.Error("Multi-container pod did not become ready: %v", err)
		os.Exit(1)
	}
	log.Success("Multi-container pod is Ready!")

	// Show all containers
	log.Command("kubectl get pod multi-container -n playground")
	out, _ = kubectl.Get("pod", "multi-container", "-n", ns)
	log.Output(out)

	fmt.Println()
	log.Command("kubectl get pod multi-container -n playground -o jsonpath='{.spec.containers[*].name}'")
	out, _ = kubectl.Get("pod", "multi-container", "-n", ns, "-o", "jsonpath={.spec.containers[*].name}")
	log.Info("Containers: %s", out)
	log.Info("All three containers share localhost. Try:")
	log.Info("  kubectl exec multi-container -n playground -c debugger -- wget -qO- http://localhost:80")
	prompt.StepPause()

	// ── Step 6: Pod logs ──────────────────────────────────────────
	log.Section("Step 6: Checking Pod Logs")
	log.Concept(
		"kubectl logs fetches stdout/stderr from a container. For multi-container\n" +
			"pods, use the -c flag to pick a container.\n" +
			"\n" +
			"Logs are your first stop when debugging. Combine with --tail, --since,\n" +
			"and --timestamps for filtering.",
	)

	log.Step("Fetching logs from nginx container...")
	nginxLogs, _ := kubectl.Logs("nginx", "", "-n", ns)
	log.Command("kubectl logs nginx -n playground")
	log.Output(firstLines(nginxLogs, 5))

	fmt.Println()
	log.Step("Fetching logs from the debugger container in multi-container pod...")
	debuggerLogs, _ := kubectl.Logs("multi-container", "debugger", "-n", ns)
	log.Command("kubectl logs multi-container -c debugger -n playground")
	log.Output(debuggerLogs)
	prompt.StepPause()

	// ── Cleanup ───────────────────────────────────────────────────
	log.Section("Cleanup: Removing Resources")
	log.Step("Deleting pods...")
	kubectl.DeleteResource("pod", "nginx", ns)
	kubectl.DeleteResource("pod", "nginx-labeled", ns)
	kubectl.DeleteResource("pod", "multi-container", ns)
	log.Success("Pods removed!")
	log.Info("(We keep the namespace for subsequent modules.)")

	// ── Summary ───────────────────────────────────────────────────
	log.Section("Summary: What You Learned")
	log.Info("  1. Namespaces isolate resources — create them before deploying")
	log.Info("  2. Pods are the atomic unit: Pending → Running → Succeeded/Failed")
	log.Info("  3. kubectl get/describe/logs/exec are your everyday tools")
	log.Info("  4. Labels enable loose coupling — use them on everything")
	log.Info("  5. Multi-container Pods share network and storage (localhost!)")
	log.Info("  6. kubectl describe shows events — your first debugging stop")
	log.Info("  7. Pods are ephemeral and immutable — deployments fix this (next!)")

	fmt.Println()
	log.Success("── Exercise 01 complete! ──")
	log.Info("Next: Module 02 — Deployments & ReplicaSets")
	log.Info("  Run: go run ./exercises/02-deployments/")
	fmt.Println()
}

// ── Helpers ────────────────────────────────────────────────────────

func splitLines(s string, max int) string {
	lines := []rune{}
	for i, r := range s {
		lines = append(lines, r)
		if r == '\n' {
			max--
			if max == 0 {
				lines = append(lines, []rune("... (truncated)")...)
				break
			}
		}
		_ = i
	}
	return string(lines)
}

func firstLine(s string) string {
	for i, r := range s {
		if r == '\n' {
			return s[:i]
		}
		_ = i
	}
	return s
}

func firstLines(s string, n int) string {
	count := 0
	for i, r := range s {
		if r == '\n' {
			count++
			if count == n {
				return s[:i] + "\n..."
			}
		}
		_ = i
	}
	return s
}
