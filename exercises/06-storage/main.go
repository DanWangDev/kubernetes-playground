// Module 06: Storage (PV/PVC) — Exercise Runner
//
// Run:
//
//	go run ./exercises/06-storage/             (automatic)
//	go run ./exercises/06-storage/ --step      (interactive step-by-step)
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/danwa/kubernetes-playground/pkg/kubectl"
	"github.com/danwa/kubernetes-playground/pkg/logger"
	"github.com/danwa/kubernetes-playground/pkg/prompt"
)

var (
	log         = logger.New("06-storage")
	manifestDir = filepath.Join("exercises", "06-storage", "manifests")
	ns          = "playground"
)

func main() {
	stepMode := flag.Bool("step", false, "Run interactively step by step")
	flag.Parse()
	if *stepMode {
		prompt.EnableStepMode()
	}

	log.Section("Module 06: Storage (PV/PVC)")
	log.Info("Domain: Static/dynamic provisioning, access modes, reclaim policies, StorageClass")
	log.Info("Duration: ~6 minutes")
	fmt.Println()

	// ── Step 1: Static PV + PVC ───────────────────────────────────
	log.Section("Step 1: Create a Static PersistentVolume")
	log.Concept(
		"A PersistentVolume (PV) is cluster-scoped storage. It's like a disk in\n" +
			"the cluster that any namespace can claim. This PV uses hostPath, which maps\n" +
			"to a directory on the node (inside the kind Docker container).",
	)

	log.Step("Creating static PV with hostPath...")
	if err := kubectl.Apply(filepath.Join(manifestDir, "01-pv-hostpath.yaml")); err != nil {
		log.Error("Failed: %v", err)
		os.Exit(1)
	}
	log.Success("PV 'pv-hostpath' created (1Gi, RWO, hostPath)!")

	log.Command("kubectl get pv")
	out, _ := kubectl.Get("pv")
	log.Output(out)
	log.Info("STATUS=Available — the PV is ready to be claimed")
	log.Info("PVs are cluster-scoped — notice there's no NAMESPACE column")
	prompt.StepPause()

	// ── Step 2: Create PVC and observe binding ────────────────────
	log.Section("Step 2: Create a PVC and Observe Binding")
	log.Concept(
		"A PersistentVolumeClaim (PVC) requests storage. It's namespaced. When a\n" +
			"PVC matches a PV (by access mode and capacity), Kubernetes binds them.\n" +
			"The binding is one-to-one — a PV can only serve one PVC.",
	)

	log.Step("Creating PVC (500Mi, RWO)...")
	if err := kubectl.Apply(filepath.Join(manifestDir, "02-pvc-claim.yaml")); err != nil {
		log.Error("Failed: %v", err)
		os.Exit(1)
	}

	log.Command("kubectl get pvc -n playground")
	out, _ = kubectl.Get("pvc", "-n", ns)
	log.Output(out)
	log.Info("STATUS=Bound — the PVC is bound to pv-hostpath")

	log.Command("kubectl get pv")
	out, _ = kubectl.Get("pv")
	log.Output(out)
	log.Info("Notice: PV STATUS changed from Available to Bound!")
	prompt.StepPause()

	// ── Step 3: Mount PVC in a Pod ────────────────────────────────
	log.Section("Step 3: Mount PVC in a Pod — Write Data")
	log.Concept(
		"Pods consume PVCs as volumes. The PVC is referenced by name in the pod spec.\n" +
			"Multiple pods on the same node can share an RWO volume (but only one pod\n" +
			"if they're on different nodes).",
	)

	log.Step("Creating pod that writes a timestamp to the PVC...")
	if err := kubectl.Apply(filepath.Join(manifestDir, "03-pod-with-pvc.yaml")); err != nil {
		log.Error("Failed: %v", err)
		os.Exit(1)
	}
	kubectl.WaitForPodReady("storage-writer", ns, kubectl.DefaultTimeout)
	log.Success("Pod 'storage-writer' is running!")

	log.Command("kubectl logs storage-writer -n playground")
	out, _ = kubectl.Logs("storage-writer", "", "-n", ns)
	log.Output(out)
	log.Success("Data written to persistent volume!")
	prompt.StepPause()

	// ── Step 4: Prove data persists ───────────────────────────────
	log.Section("Step 4: Prove Data Persistence")
	log.Concept(
		"Here's the magic of PVs: delete the pod, recreate it, and the data is\n" +
			"still there. The PVC acts as a detached disk — pods come and go, but the\n" +
			"data stays as long as the PVC exists.",
	)

	log.Step("Deleting the pod...")
	kubectl.DeleteResource("pod", "storage-writer", ns)
	log.Success("Pod deleted!")

	log.Step("Recreating the pod with the same PVC...")
	if err := kubectl.Apply(filepath.Join(manifestDir, "03-pod-with-pvc.yaml")); err != nil {
		log.Error("Failed: %v", err)
		os.Exit(1)
	}
	kubectl.WaitForPodReady("storage-writer", ns, kubectl.DefaultTimeout)

	log.Command("kubectl exec storage-writer -n playground -- cat /data/timestamp.txt")
	out, _ = kubectl.Exec("storage-writer", []string{"cat", "/data/timestamp.txt"}, "", "-n", ns)
	log.Output(out)
	log.Success("Data survived pod deletion! The PVC preserved it.")
	prompt.StepPause()

	// ── Step 5: Dynamic provisioning ──────────────────────────────
	log.Section("Step 5: Dynamic Provisioning with StorageClass")
	log.Concept(
		"With dynamic provisioning, you don't create PVs manually. You create a\n" +
			"StorageClass with a provisioner, then a PVC referencing that class.\n" +
			"The provisioner auto-creates the PV for you.",
	)

	log.Step("Creating custom StorageClass...")
	if err := kubectl.Apply(filepath.Join(manifestDir, "04-storageclass.yaml")); err != nil {
		log.Error("Failed: %v", err)
		os.Exit(1)
	}
	log.Success("StorageClass 'playground-local' created!")

	log.Step("Creating PVC with storageClassName (no pre-existing PV!)...")
	if err := kubectl.Apply(filepath.Join(manifestDir, "05-pvc-dynamic.yaml")); err != nil {
		log.Error("Failed: %v", err)
		os.Exit(1)
	}
	log.Success("PVC 'pvc-dynamic' created!")

	log.Command("kubectl get pvc -n playground")
	out, _ = kubectl.Get("pvc", "-n", ns)
	log.Output(out)
	log.Info("The StorageClass automatically provisioned a PV for this PVC!")

	log.Command("kubectl get pv | grep pvc-dynamic")
	out, _ = kubectl.Get("pv")
	// Show PVs that were dynamically created
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		if strings.Contains(line, "pvc-") {
			log.Output(line)
		}
	}
	prompt.StepPause()

	// ── Step 6: Reclaim policies ──────────────────────────────────
	log.Section("Step 6: Reclaim Policies")
	log.Concept(
		"Reclaim policies determine what happens when a PVC is deleted:\n" +
			"  Retain: PV is NOT deleted. Must be manually cleaned up.\n" +
			"  Delete (default): PV and storage are deleted automatically.\n" +
			"\n" +
			"Our static PV uses Retain. Our dynamic PVC uses Delete (from StorageClass).\n" +
			"\n" +
			"Note: We must delete the pod first — Kubernetes prevents PVC deletion\n" +
			"while a pod is still using it (PVC protection finalizer).",
	)

	log.Step("Deleting pod that uses the PVC (required before PVC deletion)...")
	kubectl.DeleteResource("pod", "storage-writer", ns)
	log.Success("Pod deleted — PVC can now be released.")

	log.Step("Deleting the dynamic PVC (should trigger PV deletion)...")
	kubectl.DeleteResource("pvc", "pvc-dynamic", ns)
	log.Success("PVC 'pvc-dynamic' deleted!")
	log.Info("The associated PV should be deleted (reclaimPolicy: Delete)")

	log.Step("Deleting static PVC (PV should remain — reclaimPolicy: Retain)...")
	log.Command("kubectl delete pvc pvc-claim -n playground")
	out, err := kubectl.Get("pvc/pvc-claim", "-n", ns, "-o", "name")
	if err == nil && strings.TrimSpace(out) != "" {
		// PVC still exists (pod was deleted above, so this should work now)
		kubectl.DeleteResource("pvc", "pvc-claim", ns)
	}
	log.Success("PVC 'pvc-claim' deleted!")
	log.Info("PV 'pv-hostpath' should remain (STATUS: Released)")
	log.Command("kubectl get pv pv-hostpath")
	out, _ = kubectl.Get("pv", "pv-hostpath")
	log.Output(out)
	log.Info("Released = PVC is gone but PV still exists. Needs manual cleanup.")
	prompt.StepPause()

	// ── Step 7: RWX explained ─────────────────────────────────────
	log.Section("Step 7: ReadWriteMany — When RWO Isn't Enough")
	log.Concept(
		"RWO means one NODE can mount, not one POD. For true multi-node shared\n" +
			"storage, you need RWX. This requires a network filesystem (NFS, Ceph, GlusterFS)\n" +
			"or a CSI driver. kind's local-path provisioner only supports RWO.\n" +
			"\n" +
			"The RWX manifest in this module demonstrates the pattern but won't bind on kind.",
	)

	log.Command("kubectl apply -f exercises/06-storage/manifests/06-readwritemany.yaml")
	kubectl.Apply(filepath.Join(manifestDir, "06-readwritemany.yaml"))
	log.Command("kubectl get pvc pvc-shared -n playground")
	out, _ = kubectl.Get("pvc", "pvc-shared", "-n", ns, "--ignore-not-found")
	log.Output(out)
	log.Warn("If STATUS=Pending, no RWX provisioner is available (expected on kind)")
	prompt.StepPause()

	// ── Cleanup ───────────────────────────────────────────────────
	log.Section("Cleanup")
	log.Step("Removing remaining resources...")
	kubectl.DeleteResource("pod", "storage-writer", ns) // already deleted, safe to retry with --ignore-not-found
	kubectl.DeleteResource("pvc", "pvc-shared", ns)
	kubectl.DeleteResource("pv", "pv-hostpath", "")     // PV is cluster-scoped, no namespace
	kubectl.DeleteResource("storageclass", "playground-local", "")
	log.Success("All storage resources cleaned up!")

	// ── Summary ───────────────────────────────────────────────────
	log.Section("Summary: What You Learned")
	log.Info("  1. PV = cluster-scoped storage; PVC = namespaced request")
	log.Info("  2. Static provisioning: admin creates PV → user creates PVC → binding")
	log.Info("  3. Dynamic provisioning: StorageClass → PVC → PV auto-created")
	log.Info("  4. Data survives Pod deletion — the PVC preserves it")
	log.Info("  5. Reclaim policies: Retain (keep PV) vs Delete (remove PV)")
	log.Info("  6. RWO = one node, RWX = many nodes (needs network storage)")
	log.Info("  7. kind's local-path provisioner = built-in dynamic provisioning")

	fmt.Println()
	log.Success("── Exercise 06 complete! ──")
	log.Info("Next: Module 07 — StatefulSets")
	log.Info("  Run: go run ./exercises/07-statefulsets/")
	fmt.Println()
}
