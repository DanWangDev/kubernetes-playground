// cluster-create creates a kind Kubernetes cluster for the playground.
// It uses the CLUSTER_NAME environment variable (default: "playground").
package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
	name := os.Getenv("CLUSTER_NAME")
	if name == "" {
		name = "playground"
	}

	fmt.Printf("Creating kind cluster %q...\n", name)
	fmt.Println("This may take 30-60 seconds on first run.")

	cmd := exec.Command("kind", "create", "cluster",
		"--name", name,
		"--wait", "2m",
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create cluster: %v\n", err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Printf("Cluster %q is ready!\n", name)
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  make cluster-status    # Verify cluster health")
	fmt.Println("  make exercise-01        # Run your first exercise")
}
