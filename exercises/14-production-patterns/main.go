// Module 14: Production Patterns — Exercise Runner
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
	log         = logger.New("14-production-patterns")
	manifestDir = filepath.Join("exercises", "14-production-patterns", "manifests")
	ns          = "playground"
)

func main() {
	stepMode := flag.Bool("step", false, "Run interactively step by step")
	flag.Parse()
	if *stepMode {
		prompt.EnableStepMode()
	}

	log.Section("Module 14: Production Patterns")
	log.Info("Domain: Affinity, anti-affinity, node selectors, taints, PDBs, topology spread, priorities")
	log.Info("Duration: ~6 minutes")
	fmt.Println()

	// ── Step 1: Pod Affinity ──────────────────────────────────────
	log.Section("Step 1: Pod Affinity — Co-locate Related Workloads")
	log.Concept(
		"Pod affinity pulls pods together on the same node. Use it to co-locate\n" +
			"related workloads (app + cache) for low latency. This deployment prefers\n" +
			"to run near pods labeled app=cache.",
	)
	kubectl.Apply(filepath.Join(manifestDir, "01-pod-affinity.yaml"))
	kubectl.WaitForDeployReady("web-with-affinity", ns, 60*time.Second)
	log.Success("Deployment with pod affinity created!")

	log.Command("kubectl get pods -n playground -l app=web-affinity -o wide")
	out, _ := kubectl.Get("pods", "-n", ns, "-l", "app=web-affinity", "-o", "wide")
	log.Output(out)
	log.Info("On a multi-node cluster, these pods would prefer nodes with app=cache pods.")
	log.Info("preferredDuringScheduling = soft constraint (scheduler tries but won't block)")
	prompt.StepPause()

	// ── Step 2: Pod Anti-Affinity ─────────────────────────────────
	log.Section("Step 2: Pod Anti-Affinity — Spread Pods Apart")
	log.Concept(
		"Pod anti-affinity pushes pods apart across nodes. Use it for high\n" +
			"availability — if one node fails, other pods on different nodes survive.",
	)
	kubectl.Apply(filepath.Join(manifestDir, "02-pod-anti-affinity.yaml"))
	kubectl.WaitForDeployReady("web-spread", ns, 60*time.Second)
	log.Success("Deployment with anti-affinity created!")

	log.Command("kubectl get pods -n playground -l app=web-spread -o wide")
	out, _ = kubectl.Get("pods", "-n", ns, "-l", "app=web-spread", "-o", "wide")
	log.Output(out)
	log.Info("On a multi-node cluster, these pods would be on different nodes.")
	log.Info("preferred (soft) anti-affinity works even on single-node kind.")
	prompt.StepPause()

	// ── Step 3: Node Selector ─────────────────────────────────────
	log.Section("Step 3: Node Selectors — Hardware Requirements")
	log.Concept(
		"nodeSelector is the simplest way to constrain pod placement. Only nodes\n" +
			"matching the label will run this pod. If no node matches, the pod stays Pending.",
	)
	kubectl.Apply(filepath.Join(manifestDir, "03-node-selector.yaml"))
	time.Sleep(2 * time.Second)

	log.Command("kubectl get pod ssd-pod -n playground")
	status, _ := kubectl.Get("pod", "ssd-pod", "-n", ns)
	log.Output(status)
	log.Warn("Pod is Pending — no node has label 'disktype: ssd' on kind!")
	log.Info("To fix: kubectl label node <name> disktype=ssd")
	log.Info("Then the pod would be scheduled to that node.")
	prompt.StepPause()

	// ── Step 4: Taints & Tolerations ──────────────────────────────
	log.Section("Step 4: Taints and Tolerations")
	log.Concept(
		"Taints REPEL pods from nodes. Tolerations ALLOW pods onto tainted nodes.\n" +
			"Think: 'This node is for GPU workloads only' (taint) and 'I need a GPU'\n" +
			"(toleration). Taints without matching tolerations = no pods scheduled.",
	)
	kubectl.Apply(filepath.Join(manifestDir, "04-taints-tolerations.yaml"))
	log.Success("Pod 'gpu-pod' created with toleration for gpu=true:NoSchedule")

	log.Command("kubectl get pod gpu-pod -n playground")
	out, _ = kubectl.Get("pod", "gpu-pod", "-n", ns)
	log.Output(out)
	log.Info("On kind (no tainted nodes), this pod runs normally.")
	log.Info("To see the pattern: kubectl taint node <name> gpu=true:NoSchedule")
	log.Info("Then only pods with matching tolerations can run there.")
	prompt.StepPause()

	// ── Step 5: Pod Disruption Budget ─────────────────────────────
	log.Section("Step 5: Pod Disruption Budget (PDB)")
	log.Concept(
		"A PDB limits voluntary disruptions (node drains, cluster upgrades).\n" +
			"It ensures a minimum number of pods stay available during maintenance.\n" +
			"minAvailable: 1 means 'keep at least 1 pod running'.",
	)
	kubectl.Apply(filepath.Join(manifestDir, "05-pdb.yaml"))
	log.Success("PDB 'web-spread-pdb' created (minAvailable: 1)!")

	log.Command("kubectl get pdb -n playground")
	out, _ = kubectl.Get("pdb", "-n", ns)
	log.Output(out)
	log.Info("When draining a node, Kubernetes respects this PDB.")
	log.Info("If only 1 pod is running and minAvailable=1, the drain is blocked.")
	prompt.StepPause()

	// ── Step 6: Topology Spread ────────────────────────────────────
	log.Section("Step 6: Topology Spread Constraints")
	log.Concept(
		"Topology spread ensures even distribution across zones/nodes:\n" +
			"  maxSkew: 1 = pod counts across zones differ by at most 1\n" +
			"  whenUnsatisfiable: ScheduleAnyway = soft constraint\n" +
			"\n" +
			"More control than podAntiAffinity — you specify exact skew tolerance.",
	)
	kubectl.Apply(filepath.Join(manifestDir, "06-topology-spread.yaml"))
	kubectl.WaitForDeployReady("web-topology", ns, 60*time.Second)
	log.Success("Deployment with topology spread created!")

	log.Command("kubectl get pods -n playground -l app=web-topology -o wide")
	out, _ = kubectl.Get("pods", "-n", ns, "-l", "app=web-topology")
	log.Output(out)
	log.Info("4 replicas with ScheduleAnyway — works even on single-node kind.")
	prompt.StepPause()

	// ── Step 7: PriorityClass ──────────────────────────────────────
	log.Section("Step 7: Priority Classes")
	log.Concept(
		"PriorityClass assigns importance to pods. Higher-priority pods can\n" +
			"preempt (evict) lower-priority pods when resources are tight. Values\n" +
			"range from -2147483648 to 1000000000.",
	)
	kubectl.Apply(filepath.Join(manifestDir, "07-priority-class.yaml"))
	log.Success("PriorityClass 'high-priority' created (value: 1000)")

	log.Command("kubectl get priorityclass")
	outPublic, _ := kubectl.Get("priorityclass")
	log.Output(outPublic)
	log.Info("Built-in classes: system-cluster-critical (2000000000), system-node-critical (2000001000)")
	log.Info("Production apps should use values below 1000000000 (user range).")
	prompt.StepPause()

	// ── Cleanup ───────────────────────────────────────────────────
	log.Section("Cleanup")
	kubectl.DeleteResource("deployment", "web-with-affinity", ns)
	kubectl.DeleteResource("deployment", "web-spread", ns)
	kubectl.DeleteResource("deployment", "web-topology", ns)
	kubectl.DeleteResource("pod", "ssd-pod", ns)
	kubectl.DeleteResource("pod", "gpu-pod", ns)
	kubectl.DeleteResource("pdb", "web-spread-pdb", ns)
	kubectl.DeleteResource("priorityclass", "high-priority", "")
	log.Success("All production pattern resources cleaned up!")

	// ── Final Summary ─────────────────────────────────────────────
	log.Section("Summary: What You Learned")
	log.Info("  1. Pod affinity co-locates related workloads (app + cache)")
	log.Info("  2. Pod anti-affinity spreads pods for high availability")
	log.Info("  3. nodeSelector is the simplest scheduling constraint")
	log.Info("  4. Taints repel pods; tolerations allow them through")
	log.Info("  5. PDBs protect availability during voluntary disruptions")
	log.Info("  6. Topology spread ensures even distribution across domains")
	log.Info("  7. PriorityClass enables preemption for critical workloads")

	fmt.Println()
	log.Success("── Exercise 14 complete! ──")
	log.Success("── Congratulations! You've completed all 14 modules! ──")
	fmt.Println()
	log.Info("You now understand Kubernetes from pods to production patterns.")
	log.Info("Project: https://github.com/DanWangDev/kubernetes-playground")
	log.Info("Run the exercises again anytime: make exercise-NN")
	fmt.Println()
}
