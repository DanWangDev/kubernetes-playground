// Module 03: Services & Networking — Exercise Runner
//
// Run:
//
//	go run ./exercises/03-services/             (automatic)
//	go run ./exercises/03-services/ --step      (interactive step-by-step)
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
	log         = logger.New("03-services")
	manifestDir = filepath.Join("exercises", "03-services", "manifests")
	ns          = "playground"
)

func main() {
	stepMode := flag.Bool("step", false, "Run interactively step by step")
	flag.Parse()
	if *stepMode {
		prompt.EnableStepMode()
	}

	log.Section("Module 03: Services & Networking")
	log.Info("Domain: ClusterIP, NodePort, DNS, Endpoints, port-forward")
	log.Info("Duration: ~6 minutes")
	fmt.Println()

	// ── Step 1: Deploy backend and ClusterIP ──────────────────────
	log.Section("Step 1: Create a Deployment and ClusterIP Service")
	log.Concept(
		"A ClusterIP Service provides a stable virtual IP and DNS name inside\n" +
			"the cluster. Clients connect to the Service, which load-balances across\n" +
			"matching Pods. Pod IPs change — the Service IP stays constant.",
	)

	log.Step("Creating nginx deployment (2 replicas)...")
	deployYAML := fmt.Sprintf(`apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-svc-demo
  namespace: %s
spec:
  replicas: 2
  selector:
    matchLabels:
      app: nginx-svc-demo
  template:
    metadata:
      labels:
        app: nginx-svc-demo
    spec:
      containers:
      - name: nginx
        image: nginx:alpine
        ports:
        - containerPort: 80
`, ns)
	if err := kubectl.ApplyString(deployYAML); err != nil {
		log.Error("Failed: %v", err)
		os.Exit(1)
	}
	kubectl.WaitForDeployReady("nginx-svc-demo", ns, kubectl.DefaultTimeout)
	log.Success("Deployment 'nginx-svc-demo' ready (2 replicas)!")

	log.Step("Creating ClusterIP Service...")
	if err := kubectl.Apply(filepath.Join(manifestDir, "01-clusterip.yaml")); err != nil {
		log.Error("Failed: %v", err)
		os.Exit(1)
	}
	log.Success("Service 'nginx-svc' created!")

	log.Command("kubectl get svc nginx-svc -n playground")
	out, _ := kubectl.Get("svc", "nginx-svc", "-n", ns)
	log.Output(out)
	log.Info("CLUSTER-IP is the virtual IP. Only reachable from inside the cluster.")

	fmt.Println()
	log.Command("kubectl get endpoints nginx-svc -n playground")
	out, _ = kubectl.Get("endpoints", "nginx-svc", "-n", ns)
	log.Output(out)
	log.Info("Endpoints list the Pod IPs the Service routes to.")
	prompt.StepPause()

	// ── Step 2: DNS resolution from debug pod ─────────────────────
	log.Section("Step 2: Test DNS Resolution from Inside the Cluster")
	log.Concept(
		"Every Service gets a DNS A record. The format is:\n" +
			"  <service>.<namespace>.svc.cluster.local\n" +
			"\n" +
			"From any pod, you can use the short name (same namespace) or full name.",
	)

	log.Step("Launching debug pod...")
	if err := kubectl.Apply(filepath.Join(manifestDir, "03-debug-pod.yaml")); err != nil {
		log.Error("Failed: %v", err)
		os.Exit(1)
	}
	kubectl.WaitForPodReady("debug", ns, kubectl.DefaultTimeout)
	log.Success("Debug pod ready!")

	// Test DNS
	log.Step("Testing DNS resolution from debug pod...")
	log.Command("kubectl exec debug -n playground -- nslookup nginx-svc")
	out, err := kubectl.Exec("debug", []string{"nslookup", "nginx-svc"}, "", "-n", ns)
	if err == nil {
		log.Output(out)
	} else {
		log.Warn("(nslookup not available in busybox, trying wget)")
	}

	// Test HTTP connectivity
	log.Step("Curling nginx-svc from debug pod...")
	out, err = kubectl.Exec("debug", []string{"wget", "-q", "-O", "-", "http://nginx-svc"}, "", "-n", ns)
	if err == nil {
		log.Output(firstLine(out))
		log.Success("Successfully reached nginx via ClusterIP DNS!")
	} else {
		log.Warn("Could not reach nginx-svc (wget may need BusyBox with wget)")
	}
	prompt.StepPause()

	// ── Step 3: NodePort ──────────────────────────────────────────
	log.Section("Step 3: NodePort — Access from Outside")
	log.Concept(
		"NodePort opens a port (30000-32767) on EVERY node. Traffic to\n" +
			"<any-node-IP>:<nodePort> reaches the Service, which routes to Pods.\n" +
			"On kind, NodePorts are mapped to your host's localhost.",
	)

	log.Step("Creating NodePort Service (port 30080)...")
	if err := kubectl.Apply(filepath.Join(manifestDir, "02-nodeport.yaml")); err != nil {
		log.Error("Failed: %v", err)
		os.Exit(1)
	}
	log.Success("NodePort Service 'nginx-nodeport' created!")

	log.Command("kubectl get svc nginx-nodeport -n playground")
	out, _ = kubectl.Get("svc", "nginx-nodeport", "-n", ns)
	log.Output(out)
	log.Info("Port 30080 on every node now routes to nginx pods.")
	log.Info("On kind: curl http://localhost:30080")
	prompt.StepPause()

	// ── Step 4: Port Forwarding ───────────────────────────────────
	log.Section("Step 4: kubectl port-forward")
	log.Concept(
		"kubectl port-forward creates a tunnel from localhost to a Pod or Service.\n" +
			"It's a development tool (not production), perfect for debugging and testing.",
	)

	log.Info("In a real terminal, you'd run:")
	log.Command("kubectl port-forward svc/nginx-svc 8080:80 -n playground")
	log.Info("Then: curl http://localhost:8080")
	log.Info("(Skipping actual port-forward since it blocks the terminal)")
	prompt.StepPause()

	// ── Step 5: Multi-Port Service ────────────────────────────────
	log.Section("Step 5: Multi-Port Service")
	log.Concept(
		"A Service can expose multiple ports with names. Named ports make intent\n" +
			"clear. Port numbers can differ between the Service and container — this\n" +
			"is port remapping (e.g., Service:443 → container:80).",
	)

	log.Step("Creating multi-port Service...")
	if err := kubectl.Apply(filepath.Join(manifestDir, "04-multi-port-service.yaml")); err != nil {
		log.Error("Failed: %v", err)
		os.Exit(1)
	}
	log.Success("Multi-port Service 'nginx-multi' created!")

	log.Command("kubectl get svc nginx-multi -n playground")
	out, _ = kubectl.Get("svc", "nginx-multi", "-n", ns)
	log.Output(out)

	log.Command("kubectl describe svc nginx-multi -n playground | grep Port:")
	out, _ = kubectl.Get("svc/nginx-multi", "-n", ns, "-o", "jsonpath={.spec.ports[*].name}")
	log.KeyValue("Port names", out)
	log.Info("http:80→80 and https:443→80 (remapped)")
	prompt.StepPause()

	// ── Step 6: Endpoints deep dive ───────────────────────────────
	log.Section("Step 6: Understanding Endpoints")
	log.Concept(
		"Endpoints (and EndpointSlices in newer K8s) list the actual Pod IPs\n" +
			"behind a Service. When a Deployment scales, Endpoints update automatically.\n" +
			"kube-proxy watches Endpoints and programs iptables/IPVS rules on each node.",
	)

	log.Command("kubectl get endpoints -n playground")
	out, _ = kubectl.Get("endpoints", "-n", ns)
	log.Output(out)

	log.Command("kubectl describe endpoints nginx-svc -n playground")
	out, _ = kubectl.Get("endpoints/nginx-svc", "-n", ns, "-o", "jsonpath={.subsets[*].addresses[*].ip}")
	log.KeyValue("Pod IPs behind nginx-svc", out)
	log.Info("These are the Pod IPs the Service routes traffic to.")
	log.Info("If you delete a Pod, its IP disappears from Endpoints and a new one appears.")
	prompt.StepPause()

	// ── Cleanup ───────────────────────────────────────────────────
	log.Section("Cleanup")
	log.Step("Removing services, deployment, and debug pod...")
	kubectl.DeleteResource("svc", "nginx-svc", ns)
	kubectl.DeleteResource("svc", "nginx-nodeport", ns)
	kubectl.DeleteResource("svc", "nginx-multi", ns)
	kubectl.DeleteResource("deployment", "nginx-svc-demo", ns)
	kubectl.DeleteResource("pod", "debug", ns)
	log.Success("All resources cleaned up!")

	// ── Summary ───────────────────────────────────────────────────
	log.Section("Summary: What You Learned")
	log.Info("  1. Services provide stable IPs and DNS for ephemeral Pods")
	log.Info("  2. ClusterIP = internal-only; NodePort = external access")
	log.Info("  3. DNS: <service>.<namespace>.svc.cluster.local")
	log.Info("  4. Endpoints track which Pod IPs are behind a Service")
	log.Info("  5. kubectl port-forward = local tunnel for debugging")
	log.Info("  6. Multi-port Services with named ports enable remapping")
	log.Info("  7. Headless Services (ClusterIP: None) → used by StatefulSets")

	fmt.Println()
	log.Success("── Exercise 03 complete! ──")
	log.Info("Next: Module 04 — ConfigMaps & Secrets")
	log.Info("  Run: go run ./exercises/04-configmaps-secrets/")
	fmt.Println()
}

func firstLine(s string) string {
	for i, r := range s {
		if r == '\n' {
			return s[:i]
		}
		_ = i
	}
	return s
}
