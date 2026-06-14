// Module 10: Autoscaling (HPA) — Exercise Runner
package main

import (
	"flag"
	"fmt"
	"path/filepath"
	"time"

	"github.com/danwa/kubernetes-playground/pkg/kubectl"
	"github.com/danwa/kubernetes-playground/pkg/logger"
	"github.com/danwa/kubernetes-playground/pkg/prompt"
)

var (
	log         = logger.New("10-autoscaling")
	manifestDir = filepath.Join("exercises", "10-autoscaling", "manifests")
	ns          = "playground"
)

func main() {
	stepMode := flag.Bool("step", false, "Run interactively step by step")
	flag.Parse()
	if *stepMode {
		prompt.EnableStepMode()
	}

	log.Section("Module 10: Autoscaling (HPA)")
	log.Info("Domain: HPA algorithm, metrics-server, CPU/memory targets, stabilization")
	log.Info("Duration: ~4 minutes")
	fmt.Println()

	// ── Step 1: Prerequisites ─────────────────────────────────────
	log.Section("Step 1: Prerequisites — Metrics Server")
	log.Concept(
		"HPA needs a metrics source. The Metrics Server collects CPU/memory metrics\n" +
			"from kubelets and exposes them via the metrics.k8s.io API.\n" +
			"\n" +
			"Install on kind (one-time setup):\n" +
			"kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml",
	)
	log.Info("(Skipping auto-install — this is a one-time cluster setup step)")
	log.Info("Without metrics-server, HPA objects are created but show <unknown> metrics.")
	log.Info("This module focuses on understanding HPA configuration.")
	prompt.StepPause()

	// ── Step 2: Create Deployment with CPU requests ───────────────
	log.Section("Step 2: Deployment with CPU Requests")
	log.Concept(
		"HPA requires resource REQUESTS on the target pods. Without requests, the\n" +
			"HPA has no baseline for percentage calculations.\n" +
			"\n" +
			"This deployment sets cpu: 100m request. The HPA will target 50% of this\n" +
			"value — meaning it scales when average CPU exceeds 50m per pod.",
	)

	log.Step("Creating deployment with CPU request (100m)...")
	kubectl.Apply(filepath.Join(manifestDir, "01-deployment-hpa.yaml"))
	kubectl.WaitForDeployReady("hpa-demo", ns, kubectl.DefaultTimeout)
	log.Success("Deployment 'hpa-demo' ready!")

	log.Command("kubectl describe deployment hpa-demo -n playground | grep -A3 Requests")
	out, _ := kubectl.Describe("deployment", "hpa-demo", "-n", ns)
	for _, line := range splitLines(out) {
		if contains(line, "cpu") || contains(line, "Requests") || contains(line, "Limits") {
			log.Output(line)
		}
	}
	log.Info("CPU request = 100m. This is the baseline for HPA percentage calculation.")
	prompt.StepPause()

	// ── Step 3: Create HPA ────────────────────────────────────────
	log.Section("Step 3: Create the HPA")
	log.Concept(
		"The HPA targets 50% average CPU utilization. It will maintain between\n" +
			"1 and 5 replicas, scaling up when CPU > 50% and down when < 50%.\n" +
			"\n" +
			"The formula: desired = ceil(current × currentMetric / targetMetric)",
	)

	log.Step("Creating HPA (target: 50 percent CPU, min: 1, max: 5)...")
	kubectl.Apply(filepath.Join(manifestDir, "02-hpa-cpu.yaml"))
	log.Success("HPA created!")

	log.Command("kubectl get hpa -n playground")
	out, _ = kubectl.Get("hpa", "-n", ns)
	log.Output(out)

	if contains(out, "<unknown>") {
		log.Warn("Metrics show <unknown> — Metrics Server is not installed.")
		log.Info("Install it to see real metrics: See Step 1 command.")
	}

	log.Command("kubectl describe hpa hpa-demo -n playground")
	out, _ = kubectl.Describe("hpa", "hpa-demo", "-n", ns)
	for _, line := range splitLines(out) {
		if contains(line, "Metrics") || contains(line, "MinReplicas") || contains(line, "MaxReplicas") || contains(line, "Deployment") {
			log.Output(line)
		}
	}
	prompt.StepPause()

	// ── Step 4: HPA in Action ─────────────────────────────────────
	log.Section("Step 4: Understanding HPA Behavior")
	log.Concept(
		"The HPA checks metrics every 15 seconds. Key behaviors:\n" +
			"  - Scale UP: fast (as soon as metrics exceed target)\n" +
			"  - Scale DOWN: slow (5-minute stabilization window by default)\n" +
			"  - Never scales beyond minReplicas or maxReplicas\n" +
			"\n" +
			"To see scaling in action, run a load generator against the deployment.",
	)

	log.Info("Load test command (in a separate terminal):")
	log.Command("kubectl run -i --tty load-generator --rm --image=busybox --restart=Never -- /bin/sh -c")
	log.Command("while true; do wget -q -O- http://hpa-demo-svc; done")
	log.Info("Then watch: kubectl get hpa hpa-demo -n playground -w")
	time.Sleep(1 * time.Second)
	prompt.StepPause()

	// ── Cleanup ───────────────────────────────────────────────────
	log.Section("Cleanup")
	kubectl.DeleteResource("hpa", "hpa-demo", ns)
	kubectl.DeleteResource("deployment", "hpa-demo", ns)
	log.Success("HPA and deployment cleaned up!")

	// ── Summary ───────────────────────────────────────────────────
	log.Section("Summary: What You Learned")
	log.Info("  1. HPA = automatic pod scaling based on metrics")
	log.Info("  2. Metrics Server must be installed (kubectl apply components.yaml)")
	log.Info("  3. Pods need resource requests for HPA to work")
	log.Info("  4. desired = ceil(current × currentMetric / targetMetric)")
	log.Info("  5. Scale up is fast; scale down has a stabilization window")
	log.Info("  6. minReplicas and maxReplicas define the scaling boundaries")

	fmt.Println()
	log.Success("── Exercise 10 complete! ──")
	log.Info("Next: Module 11 — RBAC & Security")
	log.Info("  Run: go run ./exercises/11-rbac-security/")
	fmt.Println()
}

func splitLines(s string) []string {
	var lines []string
	current := ""
	for _, r := range s {
		current += string(r)
		if r == '\n' {
			lines = append(lines, current)
			current = ""
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
