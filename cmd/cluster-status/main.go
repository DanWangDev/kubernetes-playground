// cluster-status shows the health and nodes of the playground cluster.
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

	fmt.Printf("Cluster: %s\n\n", name)

	// Cluster info
	fmt.Println("── Cluster Info ──")
	cmd := exec.Command("kubectl", "cluster-info")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "\nCluster may not be running. Run: make cluster-create\n")
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println("── Nodes ──")
	cmd = exec.Command("kubectl", "get", "nodes", "-o", "wide")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()

	fmt.Println()
	fmt.Println("── Namespaces ──")
	cmd = exec.Command("kubectl", "get", "namespaces")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}
