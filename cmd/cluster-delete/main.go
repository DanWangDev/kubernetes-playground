// cluster-delete destroys the playground kind cluster.
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

	fmt.Printf("Deleting kind cluster %q...\n", name)

	cmd := exec.Command("kind", "delete", "cluster", "--name", name)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to delete cluster: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Cluster %q deleted.\n", name)
}
