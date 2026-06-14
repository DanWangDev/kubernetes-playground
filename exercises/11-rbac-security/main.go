// Module 11: RBAC & Security — Exercise Runner
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
	log         = logger.New("11-rbac-security")
	manifestDir = filepath.Join("exercises", "11-rbac-security", "manifests")
	ns          = "playground"
)

func main() {
	stepMode := flag.Bool("step", false, "Run interactively step by step")
	flag.Parse()
	if *stepMode {
		prompt.EnableStepMode()
	}

	log.Section("Module 11: RBAC & Security")
	log.Info("Domain: ServiceAccounts, Roles, RoleBindings, SecurityContext, NetworkPolicy")
	log.Info("Duration: ~5 minutes")
	fmt.Println()

	// ── Step 1: ServiceAccount ────────────────────────────────────
	log.Section("Step 1: Create a ServiceAccount")
	log.Concept(
		"A ServiceAccount is an identity for pods. Every namespace has a 'default'\n" +
			"SA that pods use unless you specify otherwise. Custom SAs let you grant\n" +
			"specific permissions to specific workloads.",
	)
	log.Step("Creating ServiceAccount 'app-sa'...")
	kubectl.Apply(filepath.Join(manifestDir, "01-serviceaccount.yaml"))
	log.Success("ServiceAccount 'app-sa' created!")

	log.Command("kubectl get sa -n playground")
	out, _ := kubectl.Get("sa", "-n", ns)
	log.Output(out)
	prompt.StepPause()

	// ── Step 2: Role ──────────────────────────────────────────────
	log.Section("Step 2: Define a Role")
	log.Concept(
		"A Role defines permissions within a namespace. This Role allows:\n" +
			"  - get, list, watch pods\n" +
			"  - get pod logs\n" +
			"\n" +
			"Common verbs: get, list, watch, create, update, patch, delete.",
	)
	log.Step("Creating Role 'pod-reader'...")
	kubectl.Apply(filepath.Join(manifestDir, "02-role.yaml"))
	log.Success("Role 'pod-reader' created!")

	log.Command("kubectl describe role pod-reader -n playground")
	out, _ = kubectl.Describe("role", "pod-reader", "-n", ns)
	log.Output(out)
	prompt.StepPause()

	// ── Step 3: RoleBinding ───────────────────────────────────────
	log.Section("Step 3: Bind Role to ServiceAccount")
	log.Concept(
		"A RoleBinding connects a Role to a ServiceAccount. Without a binding,\n" +
			"the Role exists but nobody has those permissions. Bindings are the\n" +
			"critical 'who gets what' link in RBAC.",
	)
	log.Step("Creating RoleBinding...")
	kubectl.Apply(filepath.Join(manifestDir, "03-rolebinding.yaml"))
	log.Success("RoleBinding 'app-sa-pod-reader' created!")

	log.Command("kubectl describe rolebinding app-sa-pod-reader -n playground")
	out, _ = kubectl.Describe("rolebinding", "app-sa-pod-reader", "-n", ns)
	log.Output(out)
	log.Info("Now app-sa can list and view pods in the playground namespace!")
	prompt.StepPause()

	// ── Step 4: Test permissions ──────────────────────────────────
	log.Section("Step 4: Test RBAC Permissions")
	log.Concept(
		"kubectl auth can-i lets you check permissions without actually performing\n" +
			"the action. It's invaluable for debugging RBAC issues.",
	)
	log.Command("kubectl auth can-i list pods --as=system:serviceaccount:playground:app-sa -n playground")
	log.Info("Expected: yes — our RoleBinding grants this permission.")

	log.Command("kubectl auth can-i create deployments --as=system:serviceaccount:playground:app-sa -n playground")
	log.Info("Expected: no — our Role only covers pods, not deployments.")
	prompt.StepPause()

	// ── Step 5: SecurityContext ───────────────────────────────────
	log.Section("Step 5: Restrictive SecurityContext")
	log.Concept(
		"SecurityContext enforces pod-level security settings:\n" +
			"  runAsNonRoot: prevents running as root (UID 0)\n" +
			"  capabilities.drop: [ALL]: removes all Linux capabilities\n" +
			"  readOnlyRootFilesystem: makes root FS read-only",
	)
	log.Step("Creating pod with restrictive SecurityContext...")
	kubectl.Apply(filepath.Join(manifestDir, "04-restricted-pod.yaml"))
	kubectl.WaitForPodReady("secure-pod", ns, 30*time.Second)
	log.Success("Pod 'secure-pod' is running with restrictive security context!")

	log.Command("kubectl get pod secure-pod -n playground -o jsonpath='{.spec.securityContext}'")
	out, _ = kubectl.Get("pod/secure-pod", "-n", ns, "-o", "jsonpath={.spec.securityContext}")
	log.KeyValue("Pod SecurityContext", out)
	log.Info("runAsNonRoot + capabilities.drop: [ALL] = defense in depth")
	prompt.StepPause()

	// ── Step 6: NetworkPolicy ─────────────────────────────────────
	log.Section("Step 6: NetworkPolicy (Firewall for Pods)")
	log.Concept(
		"NetworkPolicies are pod-level firewalls. They control which pods can\n" +
			"talk to which other pods. Without a NetworkPolicy, all traffic is allowed.\n" +
			"With one, only explicitly allowed traffic flows.\n" +
			"\n" +
			"KIND'S DEFAULT CNI DOES NOT SUPPORT NETWORKPOLICY.\n" +
			"The manifest can be applied but won't be enforced on kind.",
	)
	log.Step("Applying NetworkPolicy (kind won't enforce it)...")
	kubectl.Apply(filepath.Join(manifestDir, "05-networkpolicy.yaml"))
	log.Success("NetworkPolicy 'deny-ingress' applied (conceptual on kind).")

	log.Info("For real NetworkPolicy testing, use Calico or Cilium CNI.")
	log.Info("The YAML demonstrates the pattern — it's ready for a real cluster.")
	prompt.StepPause()

	// ── Cleanup ───────────────────────────────────────────────────
	log.Section("Cleanup")
	kubectl.DeleteResource("pod", "secure-pod", ns)
	kubectl.DeleteResource("rolebinding", "app-sa-pod-reader", ns)
	kubectl.DeleteResource("role", "pod-reader", ns)
	kubectl.DeleteResource("sa", "app-sa", ns)
	kubectl.DeleteResource("networkpolicy", "deny-ingress", ns)
	log.Success("All RBAC and security resources cleaned up!")

	// ── Summary ───────────────────────────────────────────────────
	log.Section("Summary: What You Learned")
	log.Info("  1. ServiceAccounts = pod identities (not users)")
	log.Info("  2. Roles define permissions; RoleBindings grant them")
	log.Info("  3. kubectl auth can-i is your RBAC debugging tool")
	log.Info("  4. SecurityContext enforces pod-level security (best-effort)")
	log.Info("  5. NetworkPolicies are pod firewalls (need supported CNI)")
	log.Info("  6. RBAC is additive — grant only, no deny rules")

	fmt.Println()
	log.Success("── Exercise 11 complete! ──")
	log.Info("Next: Module 12 — Helm")
	log.Info("  Run: go run ./exercises/12-helm/")
	fmt.Println()
}
