// Module 13: Observability — Exercise Runner
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
	log         = logger.New("13-observability")
	manifestDir = filepath.Join("exercises", "13-observability", "manifests")
	ns          = "playground"
)

func main() {
	stepMode := flag.Bool("step", false, "Run interactively step by step")
	flag.Parse()
	if *stepMode {
		prompt.EnableStepMode()
	}

	log.Section("Module 13: Observability")
	log.Info("Domain: kubectl logs, events, describe, Prometheus, log sidecars")
	log.Info("Duration: ~5 minutes")
	fmt.Println()

	// ── Step 1: kubectl logs ──────────────────────────────────────
	log.Section("Step 1: kubectl logs — First Line of Defense")
	log.Concept(
		"When something goes wrong, kubectl logs is the first thing you check.\n" +
			"It shows stdout/stderr from containers. Key flags:\n" +
			"  --tail=N     last N lines\n" +
			"  --since=5m   last 5 minutes\n" +
			"  --previous   logs from previous (crashed) container\n" +
			"  --timestamps show when each line was written",
	)

	log.Step("Creating sample deployment...")
	kubectl.Apply(filepath.Join(manifestDir, "01-sample-app.yaml"))
	kubectl.WaitForDeployReady("obs-demo", ns, kubectl.DefaultTimeout)
	log.Success("Deployment 'obs-demo' ready!")

	log.Step("Fetching logs from one pod...")
	log.Command("kubectl logs -l app=obs-demo -n playground --tail=5")
	out, _ := kubectl.Get("pods", "-n", ns, "-l", "app=obs-demo", "-o", "jsonpath={.items[0].metadata.name}")
	podName := out
	if podName != "" {
		logs, _ := kubectl.Logs(podName, "", "-n", ns, "--tail=5")
		log.Output(logs)
	}
	log.Info("Default nginx logs show startup sequence and access attempts.")
	log.Info("Try: kubectl logs <pod> --previous (to see crash logs)")
	prompt.StepPause()

	// ── Step 2: kubectl get events ────────────────────────────────
	log.Section("Step 2: kubectl get events — What's Happening?")
	log.Concept(
		"Events record cluster activity: pod scheduling, image pulls, probe\n" +
			"results, scaling actions. They're your first stop when a pod won't\n" +
			"start or a deployment won't roll out.",
	)

	log.Command("kubectl get events -n playground --sort-by='.lastTimestamp'")
	out, _ = kubectl.Get("events", "-n", ns, "--sort-by=.lastTimestamp")
	log.Output(firstLines(out, 8))
	log.Info("Events show WHAT happened and WHY. Use them before digging into logs.")
	prompt.StepPause()

	// ── Step 3: kubectl describe ──────────────────────────────────
	log.Section("Step 3: kubectl describe — Deep Inspection")
	log.Concept(
		"kubectl describe gives the full story: Conditions, container statuses,\n" +
			"volumes, events, resource usage. It's more detailed than kubectl get\n" +
			"and includes a built-in event log at the bottom.",
	)

	log.Command("kubectl describe deployment obs-demo -n playground | head -20")
	out, _ = kubectl.Describe("deployment", "obs-demo", "-n", ns)
	log.Output(firstLines(out, 10))

	log.Info("Key sections in describe output:")
	log.Info("  Conditions: Ready? Available? Progressing?")
	log.Info("  Events: scaling decisions, image pulls")
	log.Info("  Containers: image, ports, state, readiness info")
	prompt.StepPause()

	// ── Step 4: Prometheus Architecture ────────────────────────────
	log.Section("Step 4: Prometheus + Grafana Architecture")
	log.Concept(
		"Prometheus is the standard for Kubernetes metrics. Architecture:\n" +
			"  ServiceMonitor → Prometheus → Grafana → Dashboards + Alerts\n" +
			"\n" +
			"Install via Helm (one-time):\n" +
			"  helm repo add prometheus-community https://prometheus-community.github.io/helm-charts\n" +
			"  helm install monitoring prometheus-community/kube-prometheus-stack",
	)

	log.Info("Prometheus auto-discovers pods with annotations:")
	log.Info("  prometheus.io/scrape: \"true\"")
	log.Info("  prometheus.io/port: \"80\"")
	log.Info("(Our obs-demo deployment has these annotations)")
	log.Info("After installing Prometheus, pods with these annotations are auto-scraped.")
	prompt.StepPause()

	// ── Step 5: Log Sidecar ───────────────────────────────────────
	log.Section("Step 5: Log Sidecar Pattern")
	log.Concept(
		"When an app writes logs to a file instead of stdout, use a sidecar\n" +
			"container to tail the file to stdout. This lets kubectl logs see the logs.\n" +
			"Both containers share a volume (emptyDir or PVC).",
	)

	log.Step("Deploying pod with sidecar pattern...")
	kubectl.Apply(filepath.Join(manifestDir, "02-log-sidecar.yaml"))
	kubectl.WaitForPodReady("log-sidecar-demo", ns, 30*time.Second)
	log.Success("Pod 'log-sidecar-demo' ready!")

	log.Step("Viewing logs from the app container...")
	log.Command("kubectl logs log-sidecar-demo -c app -n playground --tail=3")
	logs, _ := kubectl.Logs("log-sidecar-demo", "app", "-n", ns, "--tail=3")
	log.Output(logs)

	log.Step("Viewing logs from the log-shipper sidecar...")
	log.Command("kubectl logs log-sidecar-demo -c log-shipper -n playground")
	logs, _ = kubectl.Logs("log-sidecar-demo", "log-shipper", "-n", ns)
	log.Output(logs)

	log.Info("Each container's logs are independently accessible via -c flag.")
	log.Info("For production: use a log aggregation system (Loki, ELK, Datadog).")
	prompt.StepPause()

	// ── Cleanup ───────────────────────────────────────────────────
	log.Section("Cleanup")
	kubectl.DeleteResource("pod", "log-sidecar-demo", ns)
	kubectl.DeleteResource("deployment", "obs-demo", ns)
	log.Success("Observability resources cleaned up!")

	// ── Summary ───────────────────────────────────────────────────
	log.Section("Summary: What You Learned")
	log.Info("  1. kubectl logs = first debugging step (--tail, --previous)")
	log.Info("  2. kubectl get events = cluster activity timeline")
	log.Info("  3. kubectl describe = deep inspection with conditions + events")
	log.Info("  4. Prometheus auto-discovers pods via annotations")
	log.Info("  5. Log sidecars make file-based logs visible to kubectl")
	log.Info("  6. Structured logging (JSON) is better for production")

	fmt.Println()
	log.Success("── Exercise 13 complete! ──")
	log.Info("Next: Module 14 — Production Patterns")
	log.Info("  Run: go run ./exercises/14-production-patterns/")
	fmt.Println()
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
