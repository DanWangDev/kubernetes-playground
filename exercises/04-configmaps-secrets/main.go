// Module 04: ConfigMaps & Secrets — Exercise Runner
//
// Run:
//
//	go run ./exercises/04-configmaps-secrets/             (automatic)
//	go run ./exercises/04-configmaps-secrets/ --step      (interactive step-by-step)
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
	log         = logger.New("04-configmaps-secrets")
	manifestDir = filepath.Join("exercises", "04-configmaps-secrets", "manifests")
	ns          = "playground"
)

func main() {
	stepMode := flag.Bool("step", false, "Run interactively step by step")
	flag.Parse()
	if *stepMode {
		prompt.EnableStepMode()
	}

	log.Section("Module 04: ConfigMaps & Secrets")
	log.Info("Domain: Configuration injection, env vars, volume mounts, base64, immutability")
	log.Info("Duration: ~6 minutes")
	fmt.Println()

	// ── Step 1: Create a ConfigMap ────────────────────────────────
	log.Section("Step 1: Create a ConfigMap")
	log.Concept(
		"ConfigMaps store non-sensitive key-value configuration. Three ways to use them:\n" +
			"  valueFrom: inject a single key as an environment variable\n" +
			"  envFrom: inject ALL keys as environment variables\n" +
			"  volumes: mount each key as a file in the container",
	)

	log.Step("Creating ConfigMap from literal values...")
	if err := kubectl.Apply(filepath.Join(manifestDir, "01-configmap-env.yaml")); err != nil {
		log.Error("Failed: %v", err)
		os.Exit(1)
	}
	log.Success("ConfigMap 'app-config' created!")

	log.Command("kubectl get configmap app-config -n playground")
	out, _ := kubectl.Get("configmap", "app-config", "-n", ns)
	log.Output(out)

	log.Step("Inspecting the data...")
	log.Command("kubectl get configmap app-config -n playground -o yaml")
	out, _ = kubectl.Get("configmap/app-config", "-n", ns, "-o", "yaml")
	log.Output(firstLines(out, 12))
	prompt.StepPause()

	// ── Step 2: ConfigMap as env vars ─────────────────────────────
	log.Section("Step 2: Inject ConfigMap as Environment Variables")
	log.Concept(
		"envFrom imports all ConfigMap keys as environment variables. The env var\n" +
			"name matches the ConfigMap key. This is the simplest pattern for passing\n" +
			"configuration to your application.",
	)

	log.Step("Creating Secret and config-demo pod...")
	kubectl.Apply(filepath.Join(manifestDir, "03-secret-opaque.yaml"))
	if err := kubectl.Apply(filepath.Join(manifestDir, "05-all-env-sources.yaml")); err != nil {
		log.Error("Failed: %v", err)
		os.Exit(1)
	}
	kubectl.WaitForPodReady("config-demo", ns, kubectl.DefaultTimeout)
	log.Success("Pod 'config-demo' is running with all three injection patterns!")

	log.Step("Checking environment variables in the pod...")
	log.Command("kubectl logs config-demo -n playground")
	out, _ = kubectl.Logs("config-demo", "", "-n", ns)
	log.Output(out)
	log.Info("Notice: LOG_LEVEL came from configMapKeyRef (single key)")
	log.Info("        APP_ENV came from envFrom (bulk import)")
	log.Info("        DB_PASSWORD came from secretKeyRef (Secret)")
	prompt.StepPause()

	// ── Step 3: ConfigMap as volume ───────────────────────────────
	log.Section("Step 3: ConfigMap Mounted as Files")
	log.Concept(
		"When you mount a ConfigMap as a volume, each key becomes a file.\n" +
			"This is great for config files (YAML, JSON, TOML) that your application\n" +
			"reads at startup. The files update when the ConfigMap changes (eventually).",
	)

	log.Step("Creating file-based ConfigMap...")
	if err := kubectl.Apply(filepath.Join(manifestDir, "02-configmap-volume.yaml")); err != nil {
		log.Error("Failed: %v", err)
		os.Exit(1)
	}
	log.Success("ConfigMap 'app-files' created!")

	// Since the pod was already created, recreate it to mount the volume
	log.Step("Recreating pod to pick up volume mount...")
	kubectl.DeleteResource("pod", "config-demo", ns)
	kubectl.Apply(filepath.Join(manifestDir, "05-all-env-sources.yaml"))
	kubectl.WaitForPodReady("config-demo", ns, kubectl.DefaultTimeout)

	log.Step("Checking mounted files...")
	log.Command("kubectl exec config-demo -n playground -- cat /etc/app/config.yaml")
	out, _ = kubectl.Exec("config-demo", []string{"cat", "/etc/app/config.yaml"}, "", "-n", ns)
	log.Output(out)
	log.Info("Each ConfigMap key became a file in /etc/app/")
	prompt.StepPause()

	// ── Step 4: Secrets are base64, NOT encrypted ─────────────────
	log.Section("Step 4: Secrets — Base64, NOT Encrypted")
	log.Concept(
		"Kubernetes Secrets are base64-encoded, NOT encrypted. This is the #1\n" +
			"security misunderstanding. Anyone with kubectl access to Secrets can\n" +
			"decode them in one command.",
	)

	log.Step("Viewing the Secret (encoded)...")
	log.Command("kubectl get secret app-secret -n playground -o jsonpath='{.data.API_KEY}'")
	out, _ = kubectl.Get("secret/app-secret", "-n", ns, "-o", "jsonpath={.data.API_KEY}")
	log.Output(out)
	log.Warn("This looks encrypted, but it's just base64!")

	log.Step("Decoding the Secret...")
	log.Command("kubectl get secret app-secret -n playground -o go-template='{{range $k,$v := .data}}{{printf \"%s: \" $k}}{{if not $v}}{{$v}}{{else}}{{$v | base64decode}}{{end}}{{\"\\n\"}}{{end}}'")
	log.Info("(In a real terminal, run the command above to decode)")
	log.Info("The API_KEY decodes to: super-secret-api-key-12345")

	log.Warn("Secrets need extra protection in production:")
	log.Info("  - Encryption at rest (etcd encryption)")
	log.Info("  - RBAC: restrict Secret access")
	log.Info("  - External stores: Vault, AWS Secrets Manager, Sealed Secrets")
	prompt.StepPause()

	// ── Step 5: Immutable ConfigMaps ──────────────────────────────
	log.Section("Step 5: Immutable ConfigMaps")
	log.Concept(
		"immutable: true prevents any changes to a ConfigMap. This protects against\n" +
			"accidental edits and signals that this config should not be modified.\n" +
			"To update: delete and recreate with a new name (e.g., app-config-v2).",
	)

	log.Step("Creating immutable ConfigMap...")
	if err := kubectl.Apply(filepath.Join(manifestDir, "04-immutable-config.yaml")); err != nil {
		log.Error("Failed: %v", err)
		os.Exit(1)
	}
	log.Success("Immutable ConfigMap 'app-config-immutable' created!")

	log.Step("Attempting to edit it (will fail)...")
	// Try to patch — should fail because immutable
	patchOut, err := kubectl.Patch("configmap", "app-config-immutable", `{"data":{"VERSION":"3.0.0"}}`, ns)
	if err != nil {
		log.Warn("Edit rejected! (expected — immutable ConfigMap)")
	} else {
		log.Output(patchOut)
	}
	log.Success("Immutable config is protected from accidental edits.")
	prompt.StepPause()

	// ── Step 6: Literal vs File creation ──────────────────────────
	log.Section("Step 6: Creating ConfigMaps from kubectl")
	log.Concept(
		"You can create ConfigMaps directly from kubectl without a YAML file:\n" +
			"  From literals: kubectl create configmap --from-literal=KEY=VALUE\n" +
			"  From files: kubectl create configmap --from-file=config.yaml\n" +
			"  From env files: kubectl create configmap --from-env-file=.env",
	)

	log.Command("kubectl create configmap demo-literal --from-literal=color=blue --from-literal=size=large -n playground --dry-run=client -o yaml")
	log.Info("The --dry-run=client -o yaml pattern is great for generating YAML")
	log.Info("without actually creating the resource.")
	prompt.StepPause()

	// ── Cleanup ───────────────────────────────────────────────────
	log.Section("Cleanup")
	kubectl.DeleteResource("pod", "config-demo", ns)
	kubectl.DeleteResource("configmap", "app-config", ns)
	kubectl.DeleteResource("configmap", "app-files", ns)
	kubectl.DeleteResource("configmap", "app-config-immutable", ns)
	kubectl.DeleteResource("secret", "app-secret", ns)
	log.Success("All ConfigMaps and Secrets removed!")

	// ── Summary ───────────────────────────────────────────────────
	log.Section("Summary: What You Learned")
	log.Info("  1. ConfigMaps = non-sensitive config; Secrets = sensitive data")
	log.Info("  2. Three injection patterns: valueFrom, envFrom, volume mount")
	log.Info("  3. Volume mounts turn keys into files (auto-updates eventually)")
	log.Info("  4. Secrets are base64-ENCODED, NOT encrypted — crucial difference")
	log.Info("  5. Immutable ConfigMaps prevent accidental edits")
	log.Info("  6. kubectl create configmap --from-literal is quick for simple configs")
	log.Info("  7. Changing ConfigMaps does NOT restart pods with envFrom")

	fmt.Println()
	log.Success("── Exercise 04 complete! ──")
	log.Info("Next: Module 05 — Ingress & HTTP Routing")
	log.Info("  Run: go run ./exercises/05-ingress/")
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
