// Module 07: StatefulSets — Exercise Runner
//
// Run:
//
//	go run ./exercises/07-statefulsets/             (automatic)
//	go run ./exercises/07-statefulsets/ --step      (interactive step-by-step)
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/danwa/kubernetes-playground/pkg/kubectl"
	"github.com/danwa/kubernetes-playground/pkg/logger"
	"github.com/danwa/kubernetes-playground/pkg/prompt"
)

var (
	log         = logger.New("07-statefulsets")
	manifestDir = filepath.Join("exercises", "07-statefulsets", "manifests")
	ns          = "playground"
)

func main() {
	stepMode := flag.Bool("step", false, "Run interactively step by step")
	flag.Parse()
	if *stepMode {
		prompt.EnableStepMode()
	}

	log.Section("Module 07: StatefulSets")
	log.Info("Domain: Stable identity, ordered deployment, headless services, PVC templates")
	log.Info("Duration: ~5 minutes")
	fmt.Println()

	// ── Step 1: Headless Service ──────────────────────────────────
	log.Section("Step 1: Create a Headless Service")
	log.Concept(
		"A headless Service (clusterIP: None) is required for StatefulSets. It\n" +
			"doesn't provide load balancing — instead, DNS returns individual Pod IPs.\n" +
			"Each pod gets a predictable DNS name: <pod>.<svc>.<ns>.svc.cluster.local",
	)

	log.Step("Creating headless Service...")
	if err := kubectl.Apply(filepath.Join(manifestDir, "01-headless-service.yaml")); err != nil {
		log.Error("Failed: %v", err)
		os.Exit(1)
	}
	log.Success("Headless Service 'nginx-sts-svc' created!")

	log.Command("kubectl get svc nginx-sts-svc -n playground")
	out, _ := kubectl.Get("svc", "nginx-sts-svc", "-n", ns)
	log.Output(out)
	log.Info("CLUSTER-IP = None — this is the key difference from regular Services")
	prompt.StepPause()

	// ── Step 2: Create StatefulSet ────────────────────────────────
	log.Section("Step 2: Create a StatefulSet")
	log.Concept(
		"StatefulSet pods have sticky identities: nginx-sts-0, nginx-sts-1,\n" +
			"nginx-sts-2. They're created sequentially (0, then 1, then 2), each\n" +
			"must be Ready before the next starts. This is fundamentally different\n" +
			"from Deployments, where all pods start in parallel.",
	)

	log.Step("Creating StatefulSet (3 replicas)...")
	if err := kubectl.Apply(filepath.Join(manifestDir, "02-statefulset-basic.yaml")); err != nil {
		log.Error("Failed: %v", err)
		os.Exit(1)
	}

	// Wait for each pod sequentially to demonstrate ordering
	log.Step("Watching pods appear in order (0, 1, 2)...")
	for i := 0; i < 3; i++ {
		podName := fmt.Sprintf("nginx-sts-%d", i)
		if err := kubectl.WaitForPodReady(podName, ns, 60*time.Second); err != nil {
			log.Warn("Pod %s wait: %v", podName, err)
		} else {
			log.Success("Pod %s is Ready!", podName)
		}
	}

	log.Command("kubectl get pods -n playground -l app=nginx-sts -o wide")
	out, _ = kubectl.Get("pods", "-n", ns, "-l", "app=nginx-sts")
	log.Output(out)
	log.Info("Notice the names: nginx-sts-0, nginx-sts-1, nginx-sts-2 (not random!)")
	log.Info("Each pod has a predictable, stable identity.")
	prompt.StepPause()

	// ── Step 3: Scale down ────────────────────────────────────────
	log.Section("Step 3: Scale Down (Reverse Order)")
	log.Concept(
		"StatefulSet scales down in REVERSE order: highest ordinal first.\n" +
			"nginx-sts-2 terminates before nginx-sts-1. This protects your data —\n" +
			"the newest pod (highest ordinal) goes first.",
	)

	log.Step("Scaling down from 3 to 1 replica...")
	kubectl.Scale("statefulset", "nginx-sts", 1, ns)
	time.Sleep(2 * time.Second)

	log.Command("kubectl get pods -n playground -l app=nginx-sts")
	out, _ = kubectl.Get("pods", "-n", ns, "-l", "app=nginx-sts")
	log.Output(out)
	log.Info("Only nginx-sts-0 remains. nginx-sts-2 and nginx-sts-1 were removed (reverse order).")
	prompt.StepPause()

	// ── Step 4: Scale up ──────────────────────────────────────────
	log.Section("Step 4: Scale Up (Sequential Order)")
	log.Concept(
		"Scaling back up is sequential: 0 is already running, so 1 starts first,\n" +
			"then 2. Each pod gets the same name it had before — identity is stable.",
	)

	log.Step("Scaling back up to 3 replicas...")
	kubectl.Scale("statefulset", "nginx-sts", 3, ns)

	for i := 1; i < 3; i++ {
		podName := fmt.Sprintf("nginx-sts-%d", i)
		if err := kubectl.WaitForPodReady(podName, ns, 60*time.Second); err != nil {
			log.Warn("Pod %s wait: %v", podName, err)
		}
	}
	log.Success("All pods back — same names as before!")

	log.Command("kubectl get pods -n playground -l app=nginx-sts")
	out, _ = kubectl.Get("pods", "-n", ns, "-l", "app=nginx-sts")
	log.Output(out)
	prompt.StepPause()

	// ── Step 5: PVC Template ──────────────────────────────────────
	log.Section("Step 5: PVC Templates — Per-Pod Storage")
	log.Concept(
		"A volumeClaimTemplate creates a unique PVC for each pod:\n" +
			"  data-nginx-sts-pvc-0, data-nginx-sts-pvc-1, ...\n" +
			"\n" +
			"These PVCs survive pod deletion AND StatefulSet deletion. Storage\n" +
			"outlives pods — crucial for databases and stateful workloads.",
	)

	log.Step("Creating StatefulSet with PVC template (2 replicas)...")
	if err := kubectl.Apply(filepath.Join(manifestDir, "03-pvc-template.yaml")); err != nil {
		log.Error("Failed: %v", err)
		os.Exit(1)
	}
	kubectl.WaitForPodReady("nginx-sts-pvc-0", ns, 60*time.Second)
	kubectl.WaitForPodReady("nginx-sts-pvc-1", ns, 60*time.Second)
	log.Success("StatefulSet with PVCs ready!")

	log.Command("kubectl get pvc -n playground")
	out, _ = kubectl.Get("pvc", "-n", ns)
	log.Output(out)
	log.Info("Each pod has its own PVC (data-nginx-sts-pvc-0, data-nginx-sts-pvc-1)")
	prompt.StepPause()

	// ── Cleanup (IMPORTANT: delete STS first → pods are removed → then PVCs) ──
	log.Section("Cleanup")
	log.Step("Deleting StatefulSets (this removes pods)...")
	kubectl.DeleteResource("statefulset", "nginx-sts", ns)
	kubectl.DeleteResource("statefulset", "nginx-sts-pvc", ns)
	time.Sleep(1 * time.Second)

	log.Step("Now deleting PVCs (pods are gone, so this won't block)...")
	pvcs, _ := kubectl.Get("pvc", "-n", ns, "-o", "jsonpath={.items[*].metadata.name}")
	log.Info("Removing PVCs: %s", pvcs)
	kubectl.DeleteResource("pvc", "data-nginx-sts-pvc-0", ns)
	kubectl.DeleteResource("pvc", "data-nginx-sts-pvc-1", ns)

	log.Step("Deleting headless Service...")
	kubectl.DeleteResource("svc", "nginx-sts-svc", ns)
	log.Success("All StatefulSet resources cleaned up!")

	// ── Summary ───────────────────────────────────────────────────
	log.Section("Summary: What You Learned")
	log.Info("  1. StatefulSets provide stable pod identities (nginx-sts-0, -1, -2)")
	log.Info("  2. Headless Services (clusterIP: None) enable per-pod DNS")
	log.Info("  3. Ordered deployment: 0→1→2. Ordered teardown: 2→1→0")
	log.Info("  4. PVC templates create per-pod storage that survives deletion")
	log.Info("  5. StatefulSet = databases; Deployment = web servers")
	log.Info("  6. Delete STS before PVCs to avoid protection finalizer blocks")

	fmt.Println()
	log.Success("── Exercise 07 complete! ──")
	log.Info("Next: Module 08 — Jobs & CronJobs")
	log.Info("  Run: go run ./exercises/08-jobs-cronjobs/")
	fmt.Println()
}
