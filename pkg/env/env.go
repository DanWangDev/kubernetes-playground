// Package env provides configuration loaded from environment variables
// with sensible defaults for the Kubernetes playground.
package env

import "os"

// Config holds all playground configuration.
type Config struct {
	ClusterName   string
	Namespace     string
	Kubeconfig    string
}

// Load reads configuration from environment variables, applying defaults.
func Load() *Config {
	return &Config{
		ClusterName:   getEnv("CLUSTER_NAME", "playground"),
		Namespace:     getEnv("PLAYGROUND_NAMESPACE", "playground"),
		Kubeconfig:    os.Getenv("KUBECONFIG"),
	}
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
