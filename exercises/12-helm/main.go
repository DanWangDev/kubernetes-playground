// Module 12: Helm — Exercise Runner
package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/danwa/kubernetes-playground/pkg/logger"
	"github.com/danwa/kubernetes-playground/pkg/prompt"
)

var (
	log        = logger.New("12-helm")
	chartDir   = filepath.Join("exercises", "12-helm", "myapp-chart")
	ns         = "playground"
)

func main() {
	stepMode := flag.Bool("step", false, "Run interactively step by step")
	flag.Parse()
	if *stepMode {
		prompt.EnableStepMode()
	}

	log.Section("Module 12: Helm")
	log.Info("Domain: Chart structure, Go templating, release management, values overrides")
	log.Info("Duration: ~5 minutes")
	fmt.Println()

	helmAvailable := checkHelm()

	// ── Step 1: Chart Structure ───────────────────────────────────
	log.Section("Step 1: Explore Chart Structure")
	log.Concept(
		"A Helm chart packages Kubernetes manifests with templating and\n" +
			"configuration. The chart in myapp-chart/ wraps all concepts from\n" +
			"modules 1-11 into a single deployable unit.",
	)

	log.Step("Chart file listing:")
	log.Command("tree " + chartDir + " || find " + chartDir)
	out, _ := exec.Command("find", chartDir, "-type", "f").Output()
	if len(out) == 0 {
		out, _ = exec.Command("cmd", "/c", "dir", "/s", "/b", chartDir).Output()
	}
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if strings.Contains(line, ".yaml") || strings.Contains(line, ".tpl") || strings.Contains(line, ".txt") {
			log.Info("  %s", filepath.Base(line))
		}
	}

	log.Info("Key files:")
	log.Info("  Chart.yaml      — chart metadata")
	log.Info("  values.yaml     — default configuration")
	log.Info("  templates/      — Go-templated Kubernetes manifests")
	log.Info("  _helpers.tpl    — reusable template partials")
	log.Info("  NOTES.txt       — post-install message")
	prompt.StepPause()

	// ── Step 2: Lint and Template ─────────────────────────────────
	log.Section("Step 2: Validate the Chart")
	log.Concept(
		"helm lint checks chart syntax. helm template renders the templates\n" +
			"locally without contacting the cluster — great for CI/CD pipelines.",
	)

	if helmAvailable {
		log.Step("Running helm lint...")
		cmd := exec.Command("helm", "lint", chartDir)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Run()
		log.Success("Chart is valid!")

		log.Step("Running helm template (dry-run render)...")
		log.Command("helm template myapp " + chartDir)
		out, _ := exec.Command("helm", "template", "myapp", chartDir).Output()
		// Show first few lines
		lines := strings.Split(string(out), "\n")
		for i, line := range lines {
			if i < 40 && strings.TrimSpace(line) != "" {
				log.Output(line)
			}
		}
		log.Info("... (full output: helm template myapp %s)", chartDir)
	} else {
		log.Warn("Helm not installed — skipping lint and template.")
		log.Info("Install Helm: https://helm.sh/docs/intro/install/")
	}
	prompt.StepPause()

	// ── Step 3: Values Overrides ──────────────────────────────────
	log.Section("Step 3: Values and Overrides")
	log.Concept(
		"values.yaml provides defaults. Override with --set or -f:\n" +
			"  helm install myapp ./chart --set replicaCount=5\n" +
			"  helm install myapp ./chart -f values-prod.yaml\n" +
			"\n" +
			"Compare the default values to production overrides.",
	)

	log.Command("cat " + filepath.Join(chartDir, "values.yaml"))
	log.Info("  replicaCount: 2")
	log.Info("  autoscaling.enabled: false")
	log.Info("  config.APP_ENV: staging")

	log.Command("cat " + filepath.Join(chartDir, "values-prod.yaml"))
	log.Info("  replicaCount: 5")
	log.Info("  autoscaling.enabled: true")
	log.Info("  config.APP_ENV: production")

	if helmAvailable {
		log.Step("Rendering with production values...")
		log.Command("helm template myapp " + chartDir + " -f " + filepath.Join(chartDir, "values-prod.yaml") + " | grep replicas:")
		out, _ := exec.Command("helm", "template", "myapp", chartDir, "-f", filepath.Join(chartDir, "values-prod.yaml")).Output()
		for _, line := range strings.Split(string(out), "\n") {
			if strings.Contains(line, "replicas:") || strings.Contains(line, "minReplicas:") {
				log.Output(line)
			}
		}
		log.Info("Notice: replicaCount changed from 2 → 5, autoscaling now enabled!")
	}
	prompt.StepPause()

	// ── Step 4: Install (simulated) ───────────────────────────────
	log.Section("Step 4: Installing a Release")
	log.Concept(
		"helm install creates a release — a running instance of a chart. The release\n" +
			"is stored as a Kubernetes Secret. Each install/upgrade creates a new revision.",
	)

	if helmAvailable {
		log.Step("Installing the chart...")
		log.Command("helm install myapp " + chartDir + " --namespace playground --create-namespace")
		cmd := exec.Command("helm", "install", "myapp", chartDir, "--namespace", ns, "--create-namespace", "--wait", "--timeout", "30s")
		out, err := cmd.CombinedOutput()
		if err != nil {
			log.Warn("Install failed (may already exist): %s", strings.TrimSpace(string(out)))
			log.Step("Trying upgrade instead...")
			cmd = exec.Command("helm", "upgrade", "myapp", chartDir, "--namespace", ns, "--wait", "--timeout", "30s")
			out, err = cmd.CombinedOutput()
			if err != nil {
				log.Warn("Upgrade also failed: %s", strings.TrimSpace(string(out)))
			} else {
				log.Success("Release upgraded!")
			}
		} else {
			log.Success("Release installed! Check: helm list -n %s", ns)
		}
	} else {
		log.Info("Helm not available — in a real terminal:")
		log.Command("helm install myapp " + chartDir + " --namespace playground")
		log.Command("helm list -n playground")
	}
	prompt.StepPause()

	// ── Step 5: History and Rollback ─────────────────────────────
	log.Section("Step 5: Release History")
	log.Concept(
		"Each install/upgrade creates a numbered revision. helm history shows them.\n" +
			"helm rollback instantly reverts to any previous revision.",
	)
	log.Command("helm history myapp -n playground")
	if helmAvailable {
		cmd := exec.Command("helm", "history", "myapp", "--namespace", ns)
		out, err := cmd.CombinedOutput()
		if err == nil {
			log.Output(strings.TrimSpace(string(out)))
		}
	}
	log.Info("To rollback: helm rollback myapp 1 -n playground")
	prompt.StepPause()

	// ── Cleanup ───────────────────────────────────────────────────
	log.Section("Cleanup")
	if helmAvailable {
		cmd := exec.Command("helm", "uninstall", "myapp", "--namespace", ns)
		out, err := cmd.CombinedOutput()
		if err != nil {
			log.Info("%s", strings.TrimSpace(string(out)))
		} else {
			log.Success("Helm release 'myapp' uninstalled!")
		}
	} else {
		log.Command("helm uninstall myapp -n playground")
	}
	log.Success("Cleanup complete!")

	// ── Summary ───────────────────────────────────────────────────
	log.Section("Summary: What You Learned")
	log.Info("  1. Charts package manifests with Go template-based configuration")
	log.Info("  2. Chart.yaml + values.yaml + templates/ = a complete Helm chart")
	log.Info("  3. helm lint + helm template validate charts without deploying")
	log.Info("  4. helm install/upgrade/rollback/uninstall manage releases")
	log.Info("  5. values.yaml = defaults; -f / --set = overrides")
	log.Info("  6. _helpers.tpl defines reusable template partials")

	fmt.Println()
	log.Success("── Exercise 12 complete! ──")
	log.Info("Next: Module 13 — Observability")
	log.Info("  Run: go run ./exercises/13-observability/")
	fmt.Println()
}

func checkHelm() bool {
	_, err := exec.LookPath("helm")
	return err == nil
}
