// Package test provides integration tests for the Kubernetes playground.
// Tests require a running kind cluster (make cluster-create).
package test

import (
	"os"
	"testing"

	"github.com/danwa/kubernetes-playground/pkg/kubectl"
)

// TestMain verifies the cluster is reachable before running any tests.
func TestMain(m *testing.M) {
	if _, err := kubectl.ClusterInfo(); err != nil {
		// Skip tests gracefully if no cluster (local dev without kind)
		os.Exit(0)
	}
	os.Exit(m.Run())
}

// TestClusterReachable verifies basic kubectl connectivity.
func TestClusterReachable(t *testing.T) {
	info, err := kubectl.ClusterInfo()
	if err != nil {
		t.Skipf("Cluster not available, skipping tests: %v", err)
	}
	t.Logf("Cluster info:\n%s", info)

	nodes, err := kubectl.GetNodes()
	if err != nil {
		t.Fatalf("Failed to get nodes: %v", err)
	}
	t.Logf("Nodes:\n%s", nodes)

	if nodes == "" {
		t.Error("Expected at least one node")
	}
}
