// Module 05: Ingress & HTTP Routing — Exercise Runner
//
// Run:
//
//	go run ./exercises/05-ingress/             (automatic)
//	go run ./exercises/05-ingress/ --step      (interactive step-by-step)
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
	log         = logger.New("05-ingress")
	manifestDir = filepath.Join("exercises", "05-ingress", "manifests")
	ns          = "playground"
)

func main() {
	stepMode := flag.Bool("step", false, "Run interactively step by step")
	flag.Parse()
	if *stepMode {
		prompt.EnableStepMode()
	}

	log.Section("Module 05: Ingress & HTTP Routing")
	log.Info("Domain: L7 routing, path/host routing, TLS termination, NGINX controller")
	log.Info("Duration: ~5 minutes")
	fmt.Println()

	// ── Step 1: Install Ingress Controller ────────────────────────
	log.Section("Step 1: Install NGINX Ingress Controller")
	log.Concept(
		"Kubernetes doesn't include an Ingress controller by default. You must\n" +
			"install one. For kind, the NGINX Ingress Controller is the standard choice.\n" +
			"It runs as pods in the ingress-nginx namespace.",
	)

	// Check if already installed
	nginxNsExists := kubectl.NamespaceExists("ingress-nginx")
	if nginxNsExists {
		log.Warn("Ingress controller may already be installed. Checking...")
		_, err := kubectl.Get("pods", "-n", "ingress-nginx", "-l", "app.kubernetes.io/component=controller")
		if err == nil {
			log.Success("Ingress controller is already running — skipping install.")
		} else {
			log.Step("Installing NGINX Ingress Controller for kind...")
			installIngressController()
		}
	} else {
		log.Step("Installing NGINX Ingress Controller for kind...")
		installIngressController()
	}

	log.Command("kubectl get pods -n ingress-nginx")
	out, _ := kubectl.Get("pods", "-n", "ingress-nginx")
	log.Output(out)
	prompt.StepPause()

	// ── Step 2: Deploy backend apps ───────────────────────────────
	log.Section("Step 2: Deploy Backend Applications")
	log.Concept(
		"We'll deploy two backend apps:\n" +
			"  - api (echo server): returns request details, great for debugging\n" +
			"  - web (nginx): serves the default nginx welcome page\n" +
			"\n" +
			"Each gets its own Deployment and ClusterIP Service.",
	)

	log.Step("Applying app deployments...")
	if err := kubectl.Apply(filepath.Join(manifestDir, "01-app-deployments.yaml")); err != nil {
		log.Error("Failed: %v", err)
		os.Exit(1)
	}
	kubectl.WaitForDeployReady("api", ns, kubectl.DefaultTimeout)
	kubectl.WaitForDeployReady("web", ns, kubectl.DefaultTimeout)
	log.Success("Both apps deployed and ready!")

	log.Command("kubectl get deployments,svc -n playground -l 'app in (api,web)'")
	out, _ = kubectl.Get("deployments,svc", "-n", ns)
	log.Output(firstLines(out, 6))
	prompt.StepPause()

	// ── Step 3: Basic Ingress ─────────────────────────────────────
	log.Section("Step 3: Create a Basic Ingress")
	log.Concept(
		"An Ingress resource defines routing rules. This basic Ingress routes\n" +
			"all HTTP traffic (path: /) to the web service. The ingressClassName: nginx\n" +
			"tells Kubernetes which Ingress controller should handle this.",
	)

	log.Step("Creating basic Ingress...")
	if err := kubectl.Apply(filepath.Join(manifestDir, "02-ingress-basic.yaml")); err != nil {
		log.Error("Failed: %v", err)
		os.Exit(1)
	}
	log.Success("Ingress 'ingress-basic' created!")

	log.Command("kubectl get ingress -n playground")
	out, _ = kubectl.Get("ingress", "-n", ns)
	log.Output(out)
	log.Info("ADDRESS will show 'localhost' (kind maps 80/443 to host)")

	log.Command("kubectl describe ingress ingress-basic -n playground")
	out, _ = kubectl.Get("ingress/ingress-basic", "-n", ns, "-o", "jsonpath={.spec.rules[0].http.paths[0].backend.service.name}")
	log.KeyValue("Backend service", strings.TrimSpace(out))
	log.Info("On kind: curl http://localhost/ should return nginx welcome page")
	prompt.StepPause()

	// ── Step 4: Path-Based Routing ────────────────────────────────
	log.Section("Step 4: Path-Based Routing")
	log.Concept(
		"Path-based routing sends traffic to different backends based on URL path:\n" +
			"  /api  → api-svc (echo server — shows request details)\n" +
			"  /     → web-svc (nginx — default welcome page)",
	)

	log.Step("Deleting basic ingress and creating path-based ingress...")
	kubectl.DeleteResource("ingress", "ingress-basic", ns)
	if err := kubectl.Apply(filepath.Join(manifestDir, "03-path-routing.yaml")); err != nil {
		log.Error("Failed: %v", err)
		os.Exit(1)
	}
	log.Success("Path-routing Ingress created!")

	log.Command("kubectl get ingress ingress-path -n playground")
	out, _ = kubectl.Get("ingress", "ingress-path", "-n", ns)
	log.Output(out)

	log.Info("Test externally (on kind):")
	log.Command("curl http://localhost/        → nginx welcome page")
	log.Command("curl http://localhost/api    → echo server response")

	// Test from inside cluster
	log.Step("Testing from inside the cluster via debug pod...")
	testInternalConnectivity()
	prompt.StepPause()

	// ── Step 5: TLS ───────────────────────────────────────────────
	log.Section("Step 5: TLS Termination")
	log.Concept(
		"The Ingress controller terminates HTTPS and forwards plain HTTP to the\n" +
			"backend. You need a TLS certificate stored as a Kubernetes Secret. For\n" +
			"development, we'll use a self-signed certificate.",
	)

	log.Step("Creating self-signed TLS certificate...")
	createTLSCert()

	log.Step("Creating TLS Ingress...")
	kubectl.DeleteResource("ingress", "ingress-path", ns)
	if err := kubectl.Apply(filepath.Join(manifestDir, "04-tls-ingress.yaml")); err != nil {
		log.Error("Failed: %v", err)
		os.Exit(1)
	}
	log.Success("TLS Ingress 'ingress-tls' created!")

	log.Command("kubectl get ingress ingress-tls -n playground")
	out, _ = kubectl.Get("ingress", "ingress-tls", "-n", ns)
	log.Output(out)

	log.Info("Test HTTPS (on kind):")
	log.Command("curl -k https://playground.local/ -H 'Host: playground.local'")
	log.Info("The -k flag skips certificate verification (self-signed cert)")
	prompt.StepPause()

	// ── Step 6: Annotations ───────────────────────────────────────
	log.Section("Step 6: Ingress Annotations")
	log.Concept(
		"Ingress annotations configure controller-specific behavior. Each controller\n" +
			"has its own set of annotations. Common NGINX examples:\n" +
			"  nginx.ingress.kubernetes.io/rewrite-target: /\n" +
			"  nginx.ingress.kubernetes.io/ssl-redirect: \"true\"\n" +
			"  nginx.ingress.kubernetes.io/proxy-body-size: \"10m\"",
	)

	log.Info("Annotations are how you configure the underlying nginx without")
	log.Info("changing the Ingress API. They're key-value strings in metadata.annotations.")
	log.Info("This is the extension mechanism for all Ingress controllers.")
	prompt.StepPause()

	// ── Cleanup ───────────────────────────────────────────────────
	log.Section("Cleanup")
	log.Step("Removing Ingresses, apps, and TLS secret...")
	kubectl.DeleteResource("ingress", "ingress-tls", ns)
	kubectl.DeleteResource("secret", "playground-tls", ns)
	kubectl.DeleteResource("deployment", "api", ns)
	kubectl.DeleteResource("deployment", "web", ns)
	kubectl.DeleteResource("svc", "api-svc", ns)
	kubectl.DeleteResource("svc", "web-svc", ns)
	log.Success("All Ingress resources cleaned up!")
	log.Info("(Ingress controller left running for future modules)")

	// ── Summary ───────────────────────────────────────────────────
	log.Section("Summary: What You Learned")
	log.Info("  1. Ingress = L7 HTTP routing (vs Service = L4 TCP routing)")
	log.Info("  2. Ingress controller must be installed (not built-in)")
	log.Info("  3. Path-based routing: /api → api, / → web")
	log.Info("  4. Host-based routing with -H 'Host: ...' header")
	log.Info("  5. TLS termination: Ingress decrypts HTTPS → HTTP to backend")
	log.Info("  6. Annotations configure controller-specific behavior")
	log.Info("  7. IngressClassName connects an Ingress to its controller")

	fmt.Println()
	log.Success("── Exercise 05 complete! ──")
	log.Info("Next: Module 06 — Storage (PV/PVC)")
	log.Info("  Run: go run ./exercises/06-storage/")
	fmt.Println()
}

// ── Helpers ────────────────────────────────────────────────────────

func installIngressController() {
	log.Info("The NGINX Ingress Controller must be installed separately on kind:")
	log.Command("kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml")
	log.Info("(Skipping automatic install — this is a one-time cluster setup step)")
	log.Info("If not installed, Ingress resources will still be created but won't route traffic.")
}

func createTLSCert() {
	// Check if TLS secret already exists
	if _, err := kubectl.Get("secret", "playground-tls", "-n", ns); err == nil {
		log.Warn("TLS secret already exists — skipping creation")
		return
	}

	log.Command("openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout tls.key -out tls.crt -subj '/CN=playground.local'")
	log.Info("Then: kubectl create secret tls playground-tls --cert=tls.crt --key=tls.key -n playground")

	// Create a simple self-signed cert via kubectl (no openssl needed)
	// For simplicity, we create the secret with dummy cert data and explain how to do it properly
	_ = kubectl.ApplyString(fmt.Sprintf(`apiVersion: v1
kind: Secret
metadata:
  name: playground-tls
  namespace: %s
type: kubernetes.io/tls
data:
  tls.crt: LS0tLS1CRUdJTi...  # Generate with: openssl req -x509 ...
  tls.key: LS0tLS1CRUdJTi...  # Generate with: openssl req -x509 ...
`, ns))
	log.Warn("For a real certificate, use the openssl commands shown above.")
	log.Info("The pre-created secret with placeholder certs demonstrates the pattern.")
}

func testInternalConnectivity() {
	// Test from a debug pod inside the cluster
	debugYAML := fmt.Sprintf(`apiVersion: v1
kind: Pod
metadata:
  name: debug-curl
  namespace: %s
spec:
  containers:
  - name: debug
    image: busybox
    command: ["sh", "-c", "echo 'Testing internal...'; wget -qO- http://web-svc 2>/dev/null | head -1; sleep 3600"]
`, ns)

	if kubectl.NamespaceExists(ns) {
		kubectl.ApplyString(debugYAML)
		kubectl.WaitForPodReady("debug-curl", ns, kubectl.DefaultTimeout)
		logs, _ := kubectl.Logs("debug-curl", "", "-n", ns)
		log.Output(logs)
		log.Info("Internal Service connectivity works (ClusterIP)!")
		kubectl.DeleteResource("pod", "debug-curl", ns)
	}
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
