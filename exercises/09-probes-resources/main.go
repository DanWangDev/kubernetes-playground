// Module 09: Probes & Resource Management — Exercise Runner
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
	log         = logger.New("09-probes-resources")
	manifestDir = filepath.Join("exercises", "09-probes-resources", "manifests")
	ns          = "playground"
)

func main() {
	stepMode := flag.Bool("step", false, "Run interactively step by step")
	flag.Parse()
	if *stepMode {
		prompt.EnableStepMode()
	}

	log.Section("Module 09: Probes & Resource Management")
	log.Info("Domain: Liveness/readiness/startup probes, requests/limits, QoS classes")
	log.Info("Duration: ~5 minutes")
	fmt.Println()

	// ── Step 1: Readiness Probe ──────────────────────────────────
	log.Section("Step 1: Readiness Probe")
	log.Concept(
		"A readiness probe determines if a pod is ready to receive traffic.\n" +
			"If it fails, the pod is removed from Service endpoints but NOT restarted.\n" +
			"Use readiness probes for dependency checks (database, cache, upstream).",
	)
	log.Step("Creating pod with HTTP readiness probe...")
	kubectl.Apply(filepath.Join(manifestDir, "01-readiness-probe.yaml"))
	kubectl.WaitForPodReady("nginx-ready", ns, 30*time.Second)
	log.Success("Pod 'nginx-ready' is Ready — readiness probe passed!")

	log.Command("kubectl describe pod nginx-ready -n playground | grep -A5 Readiness")
	out, _ := kubectl.Describe("pod", "nginx-ready", "-n", ns)
	for _, line := range splitLines(out) {
		if contains(line, "Readiness") || contains(line, "Ready") || contains(line, "State") {
			log.Output(line)
		}
	}
	prompt.StepPause()

	// ── Step 2: Liveness Probe ───────────────────────────────────
	log.Section("Step 2: Liveness Probe")
	log.Concept(
		"A liveness probe detects deadlocks. If it fails, Kubernetes restarts\n" +
			"the container. Use liveness probes ONLY for deadlock detection — your\n" +
			"app crashing already triggers a restart without one.\n" +
			"\n" +
			"This pod creates /tmp/healthy, then deletes it after 30s. The liveness\n" +
			"probe (cat /tmp/healthy) will then fail, causing a restart.",
	)
	log.Step("Creating pod with liveness probe (file removed after 30s)...")
	kubectl.Apply(filepath.Join(manifestDir, "02-liveness-probe.yaml"))
	time.Sleep(5 * time.Second) // Let it start

	log.Command("kubectl get pod liveness-demo -n playground")
	out, _ = kubectl.Get("pod", "liveness-demo", "-n", ns)
	log.Output(out)
	log.Info("The pod is Running now. In ~30 seconds, /tmp/healthy will be deleted,")
	log.Info("the liveness probe will fail, and Kubernetes will restart the container.")
	log.Info("Watch: kubectl get pod liveness-demo -n playground -w")
	prompt.StepPause()

	// ── Step 3: Startup Probe ────────────────────────────────────
	log.Section("Step 3: Startup Probe")
	log.Concept(
		"A startup probe protects slow-starting containers. While it's running,\n" +
			"liveness and readiness probes are DISABLED. This prevents a slow app\n" +
			"from being killed by an aggressive liveness probe before it finishes\n" +
			"initialization.",
	)
	log.Step("Creating pod with startup probe (10s startup delay)...")
	kubectl.Apply(filepath.Join(manifestDir, "03-startup-probe.yaml"))
	kubectl.WaitForPodReady("slow-start", ns, 30*time.Second)
	log.Success("Pod 'slow-start' is Ready! (startup took ~10s)")

	log.Command("kubectl describe pod slow-start -n playground | grep -E '(Startup|Liveness|Readiness):'")
	out, _ = kubectl.Describe("pod", "slow-start", "-n", ns)
	for _, line := range splitLines(out) {
		if contains(line, "Startup") || contains(line, "Liveness") || contains(line, "Readiness") {
			log.Output(line)
		}
	}
	prompt.StepPause()

	// ── Step 4: Resource Requests & Limits ───────────────────────
	log.Section("Step 4: Resource Requests and Limits")
	log.Concept(
		"requests = minimum guaranteed resources (scheduler uses this)\n" +
			"limits = maximum allowed (CPU throttled, memory OOMKilled)\n" +
			"\n" +
			"Always set requests. Without them, the scheduler can't make good decisions.",
	)

	log.Step("Creating pod with Guaranteed QoS (requests = limits)...")
	kubectl.Apply(filepath.Join(manifestDir, "04-resource-requests.yaml"))
	kubectl.WaitForPodReady("guaranteed-pod", ns, 30*time.Second)

	log.Command("kubectl get pod guaranteed-pod -n playground -o jsonpath='{.spec.containers[0].resources}'")
	out, _ = kubectl.Get("pod/guaranteed-pod", "-n", ns, "-o", "jsonpath={.spec.containers[0].resources}")
	log.KeyValue("Resources", out)

	qos, _ := kubectl.Get("pod/guaranteed-pod", "-n", ns, "-o", "jsonpath={.status.qosClass}")
	log.KeyValue("QoS Class", qos)

	log.Step("Creating pod with Burstable QoS (requests < limits)...")
	kubectl.Apply(filepath.Join(manifestDir, "05-resource-limits.yaml"))
	kubectl.WaitForPodReady("burstable-pod", ns, 30*time.Second)
	qos, _ = kubectl.Get("pod/burstable-pod", "-n", ns, "-o", "jsonpath={.status.qosClass}")
	log.KeyValue("QoS Class (burstable)", qos)

	log.Info("Eviction order: BestEffort → Burstable → Guaranteed")
	prompt.StepPause()

	// ── Cleanup ───────────────────────────────────────────────────
	log.Section("Cleanup")
	kubectl.DeleteResource("pod", "nginx-ready", ns)
	kubectl.DeleteResource("pod", "liveness-demo", ns)
	kubectl.DeleteResource("pod", "slow-start", ns)
	kubectl.DeleteResource("pod", "guaranteed-pod", ns)
	kubectl.DeleteResource("pod", "burstable-pod", ns)
	log.Success("All probe and resource pods cleaned up!")

	// ── Summary ───────────────────────────────────────────────────
	log.Section("Summary: What You Learned")
	log.Info("  1. Readiness = can you serve? (removes from endpoints, no restart)")
	log.Info("  2. Liveness = are you deadlocked? (restarts the container)")
	log.Info("  3. Startup = give slow apps time before probes kick in")
	log.Info("  4. httpGet, exec, and tcpSocket are the three probe mechanisms")
	log.Info("  5. requests = minimum; limits = maximum (CPU throttled, memory killed)")
	log.Info("  6. QoS: Guaranteed > Burstable > BestEffort (eviction priority)")

	fmt.Println()
	log.Success("── Exercise 09 complete! ──")
	log.Info("Next: Module 10 — Autoscaling (HPA)")
	log.Info("  Run: go run ./exercises/10-autoscaling/")
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
