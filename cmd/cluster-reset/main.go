// cluster-reset deletes the playground cluster and creates a fresh one.
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

	// Delete
	fmt.Printf("Deleting cluster %q...\n", name)
	deleteCmd := exec.Command("kind", "delete", "cluster", "--name", name)
	deleteCmd.Stdout = os.Stdout
	deleteCmd.Stderr = os.Stderr
	deleteCmd.Run() // Ignore errors (cluster might not exist)

	// Create
	fmt.Printf("\nCreating fresh cluster %q...\n", name)
	createCmd := exec.Command("kind", "create", "cluster",
		"--name", name,
		"--wait", "2m",
	)
	createCmd.Stdout = os.Stdout
	createCmd.Stderr = os.Stderr
	createCmd.Stdin = os.Stdin

	if err := createCmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create cluster: %v\n", err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Printf("Cluster %q reset complete!\n", name)
}
