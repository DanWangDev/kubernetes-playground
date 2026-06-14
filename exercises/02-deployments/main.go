// Module 02: Deployments & ReplicaSets — Exercise Runner
//
// Run:
//
//	go run ./exercises/02-deployments/             (automatic)
//	go run ./exercises/02-deployments/ --step      (interactive step-by-step)
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/danwa/kubernetes-playground/pkg/kubectl"
	"github.com/danwa/kubernetes-playground/pkg/logger"
	"github.com/danwa/kubernetes-playground/pkg/prompt"
)

var (
	log         = logger.New("02-deployments")
	manifestDir = filepath.Join("exercises", "02-deployments", "manifests")
	ns          = "playground"
)

func main() {
	stepMode := flag.Bool("step", false, "Run interactively step by step")
	flag.Parse()
	if *stepMode {
		prompt.EnableStepMode()
	}

	log.Section("Module 02: Deployments & ReplicaSets")
	log.Info("Domain: Declarative management, rolling updates, rollbacks, ReplicaSets")
	log.Info("Duration: ~7 minutes")
	fmt.Println()

	// ── Step 1: Create a Deployment ───────────────────────────────
	log.Section("Step 1: Create Your First Deployment")
	log.Concept(
		"A Deployment manages a ReplicaSet, which manages Pods. Instead of\n" +
			"creating Pods one-by-one, you declare the desired count and template.\n" +
			"Kubernetes continuously reconciles reality toward that desired state.\n" +
			"\n" +
			"Ownership chain: Deployment → ReplicaSet → Pod",
	)

	log.Step("Applying deployment manifest (3 replicas)...")
	manifest := filepath.Join(manifestDir, "01-deployment-basic.yaml")
	if err := kubectl.Apply(manifest); err != nil {
		log.Error("Failed to create deployment: %v", err)
		os.Exit(1)
	}
	log.Success("Deployment 'nginx-deploy' created!")

	log.Step("Waiting for all replicas to be available...")
	if err := kubectl.WaitForDeployReady("nginx-deploy", ns, kubectl.DefaultTimeout); err != nil {
		log.Error("Deployment did not become ready: %v", err)
		os.Exit(1)
	}
	log.Success("All 3 replicas are Ready!")

	log.Command("kubectl get deployment nginx-deploy -n playground")
	out, _ := kubectl.Get("deployment", "nginx-deploy", "-n", ns)
	log.Output(out)

	fmt.Println()
	log.Command("kubectl get replicasets -n playground -l app=nginx")
	out, _ = kubectl.Get("replicasets", "-n", ns, "-l", "app=nginx")
	log.Output(out)
	log.Info("Notice: the ReplicaSet name includes a pod-template-hash suffix")

	fmt.Println()
	log.Command("kubectl get pods -n playground -l app=nginx")
	out, _ = kubectl.Get("pods", "-n", ns, "-l", "app=nginx")
	log.Output(out)
	log.Info("All 3 pods were created by the ReplicaSet, not directly by us.")
	prompt.StepPause()

	// ── Step 2: Scale ─────────────────────────────────────────────
	log.Section("Step 2: Scaling Up and Down")
	log.Concept(
		"Scaling changes the replica count. This does NOT create a new\n" +
			"ReplicaSet — only pod template changes trigger new ReplicaSets.\n" +
			"\n" +
			"  Imperial: kubectl scale deployment/name --replicas=N\n" +
			"  Declarative: edit replicas: N in YAML and kubectl apply",
	)

	log.Step("Scaling up to 5 replicas...")
	if err := kubectl.Scale("deployment", "nginx-deploy", 5, ns); err != nil {
		log.Warn("Scale failed: %v", err)
	}
	time.Sleep(1 * time.Second)
	log.Success("Scaled to 5!")
	log.Command("kubectl get pods -n playground -l app=nginx -o wide")
	out, _ = kubectl.Get("pods", "-n", ns, "-l", "app=nginx", "-o", "wide")
	log.Output(out)

	log.Step("Scaling back down to 2 replicas...")
	kubectl.Scale("deployment", "nginx-deploy", 2, ns)
	time.Sleep(500 * time.Millisecond)
	log.Success("Scaled back to 2.")
	log.Command("kubectl get pods -n playground -l app=nginx")
	out, _ = kubectl.Get("pods", "-n", ns, "-l", "app=nginx")
	log.Output(out)
	prompt.StepPause()

	// ── Step 3: Rolling Update ────────────────────────────────────
	log.Section("Step 3: Rolling Update")
	log.Concept(
		"When you change the Pod template (e.g., update the image), Kubernetes\n" +
			"performs a rolling update with ZERO downtime:\n" +
			"  1. Create new ReplicaSet with updated Pod template\n" +
			"  2. Gradually scale new up, old down\n" +
			"  3. Old ReplicaSet kept at 0 for instant rollback",
	)

	log.Step("Updating image to nginx:1.25-alpine to trigger rolling update...")
	log.Command("kubectl set image deployment/nginx-deploy nginx=nginx:1.25-alpine -n playground")
	kubectl.SetImage("deployment/nginx-deploy", "nginx=nginx:1.25-alpine", ns)
	log.Success("Image update applied!")

	log.Step("Watching rollout status...")
	if err := kubectl.RolloutStatus("deployment", "nginx-deploy", "-n", ns); err != nil {
		log.Warn("(timeout checking rollout status)")
	}
	log.Success("Rollout complete!")

	log.Command("kubectl get replicasets -n playground -l app=nginx")
	out, _ = kubectl.Get("replicasets", "-n", ns, "-l", "app=nginx")
	log.Output(out)
	log.Info("Two ReplicaSets: old (scaled to 0) and new (active)")
	prompt.StepPause()

	// ── Step 4: Rollout History & Rollback ────────────────────────
	log.Section("Step 4: Rollout History and Rollback")
	log.Concept(
		"Every template change creates a revision. kubectl rollout history shows\n" +
			"them. kubectl rollout undo performs an instant rollback by scaling the\n" +
			"previous ReplicaSet back up.",
	)

	log.Command("kubectl rollout history deployment/nginx-deploy -n playground")
	hist, _ := kubectl.RolloutHistory("deployment/nginx-deploy", "-n", ns)
	log.Output(hist)

	log.Step("Rolling back to revision 1...")
	kubectl.RolloutUndo("deployment", "nginx-deploy", "-n", ns)
	kubectl.RolloutStatus("deployment", "nginx-deploy", "-n", ns)
	log.Success("Rollback complete — back to nginx:alpine!")

	log.Command("kubectl get replicasets -n playground -l app=nginx")
	out, _ = kubectl.Get("replicasets", "-n", ns, "-l", "app=nginx")
	log.Output(out)
	log.Info("The original ReplicaSet is active again at full scale.")
	prompt.StepPause()

	// ── Step 5: Strategy Demo ─────────────────────────────────────
	log.Section("Step 5: Tuned RollingUpdate Strategy")
	log.Concept(
		"maxSurge and maxUnavailable control rollout pace:\n" +
			"  maxSurge=1: at most 1 extra pod during update\n" +
			"  maxUnavailable=1: at most 1 pod down during update",
	)

	log.Step("Creating deployment with explicit strategy (4 replicas)...")
	manifest = filepath.Join(manifestDir, "02-deployment-strategy.yaml")
	if err := kubectl.Apply(manifest); err != nil {
		log.Error("Failed to create deployment: %v", err)
		os.Exit(1)
	}
	kubectl.WaitForDeployReady("nginx-tuned", ns, kubectl.DefaultTimeout)
	log.Success("Deployment 'nginx-tuned' ready with 4 replicas!")

	ready, total, _ := kubectl.GetDeployReplicas("nginx-tuned", ns)
	log.KeyValue("Replicas", fmt.Sprintf("%s/%s", ready, total))
	prompt.StepPause()

	// ── Step 6: Pod Resilience ────────────────────────────────────
	log.Section("Step 6: Self-Healing — Delete a Pod")
	log.Concept(
		"Deployments guarantee availability. Delete any Pod, and the ReplicaSet\n" +
			"immediately creates a replacement. Watch this happen in real time.",
	)

	log.Command("kubectl get pods -n playground -l app=nginx")
	out, _ = kubectl.Get("pods", "-n", ns, "-l", "app=nginx")
	log.Output(firstLines(out, 3))

	podName, err := kubectl.GetFirstPodName("app=nginx", ns)
	if err == nil && podName != "" {
		log.Step("Deleting pod %s...", podName)
		kubectl.DeleteResource("pod", podName, ns)
		log.Success("Pod %s deleted!", podName)

		time.Sleep(3 * time.Second)
		log.Command("kubectl get pods -n playground -l app=nginx")
		out, _ = kubectl.Get("pods", "-n", ns, "-l", "app=nginx")
		log.Output(out)
		log.Info("The ReplicaSet immediately created a replacement!")
		log.Info("Look at the AGE column — a brand-new pod appeared.")
	}
	prompt.StepPause()

	// ── Cleanup ───────────────────────────────────────────────────
	log.Section("Cleanup")
	log.Step("Deleting deployments (removes ReplicaSets AND Pods)...")
	kubectl.DeleteResource("deployment", "nginx-deploy", ns)
	kubectl.DeleteResource("deployment", "nginx-tuned", ns)
	log.Success("All deployments cleaned up!")

	log.Command("kubectl get pods -n playground")
	out, _ = kubectl.Get("pods", "-n", ns)
	log.Output(out)
	log.Info("(should be empty — all pods cleaned up)")
	fmt.Println()

	// ── Summary ───────────────────────────────────────────────────
	log.Section("Summary: What You Learned")
	log.Info("  1. Deployment → ReplicaSet → Pod is a 3-tier ownership chain")
	log.Info("  2. Declare desired state in YAML; Kubernetes reconciles it")
	log.Info("  3. Rolling updates replace pods gradually with zero downtime")
	log.Info("  4. Old ReplicaSets are kept for instant rollback (rollout undo)")
	log.Info("  5. kubectl scale changes count; template change triggers new RS")
	log.Info("  6. Deleting a pod proves self-healing — ReplicaSet replaces it")
	log.Info("  7. maxSurge/maxUnavailable control rollout pace")

	fmt.Println()
	log.Success("── Exercise 02 complete! ──")
	log.Info("Next: Module 03 — Services & Networking")
	log.Info("  Run: go run ./exercises/03-services/")
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

// Ensure strings import is used
var _ = strings.TrimSpace
