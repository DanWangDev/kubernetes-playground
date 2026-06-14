// Module 08: Jobs & CronJobs — Exercise Runner
//
// Run:
//
//	go run ./exercises/08-jobs-cronjobs/             (automatic)
//	go run ./exercises/08-jobs-cronjobs/ --step      (interactive step-by-step)
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/danwa/kubernetes-playground/pkg/kubectl"
	"github.com/danwa/kubernetes-playground/pkg/logger"
	"github.com/danwa/kubernetes-playground/pkg/prompt"
)

var (
	log         = logger.New("08-jobs-cronjobs")
	manifestDir = filepath.Join("exercises", "08-jobs-cronjobs", "manifests")
	ns          = "playground"
)

func main() {
	stepMode := flag.Bool("step", false, "Run interactively step by step")
	flag.Parse()
	if *stepMode {
		prompt.EnableStepMode()
	}

	log.Section("Module 08: Jobs & CronJobs")
	log.Info("Domain: One-shot tasks, parallel workers, CronJobs, backoff, TTL")
	log.Info("Duration: ~5 minutes")
	fmt.Println()

	// ── Step 1: Simple Job ────────────────────────────────────────
	log.Section("Step 1: Simple One-Shot Job")
	log.Concept(
		"A Job creates Pods that run to completion. Unlike Deployments (which keep\n" +
			"pods running forever), Jobs are for finite tasks: migrations, backups,\n" +
			"batch processing. After completion, pods stick around for logs.",
	)

	log.Step("Creating simple Job...")
	if err := kubectl.Apply(filepath.Join(manifestDir, "01-job-simple.yaml")); err != nil {
		log.Error("Failed: %v", err)
		os.Exit(1)
	}

	log.Step("Waiting for Job to complete...")
	// Use kubectl wait for job condition
	if err := kubectl.Wait("job", "hello-job", "condition=Complete", "30s"); err != nil {
		log.Warn("Job wait: %v (may still be running)", err)
	}

	log.Command("kubectl get job hello-job -n playground")
	out, _ := kubectl.Get("job", "hello-job", "-n", ns)
	log.Output(out)

	log.Command("kubectl get pods -n playground -l job-name=hello-job")
	out, _ = kubectl.Get("pods", "-n", ns, "-l", "job-name=hello-job")
	log.Output(out)

	logs, _ := kubectl.Logs("hello-job", "", "-n", ns, "--tail=5")
	log.Command("kubectl logs job/hello-job -n playground")
	log.Output(logs)
	prompt.StepPause()

	// ── Step 2: Parallel Job ──────────────────────────────────────
	log.Section("Step 2: Parallel Job (5 completions, 2 workers)")
	log.Concept(
		"Parallel Jobs use multiple workers to process items faster:\n" +
			"  completions=5: need 5 successful pod runs\n" +
			"  parallelism=2: at most 2 pods run concurrently\n" +
			"\n" +
			"Wall time ≈ ceil(completions/parallelism) × per-task-time",
	)

	log.Step("Creating parallel Job...")
	if err := kubectl.Apply(filepath.Join(manifestDir, "02-job-parallel.yaml")); err != nil {
		log.Error("Failed: %v", err)
		os.Exit(1)
	}

	log.Step("Waiting for parallel Job...")
	kubectl.Wait("job", "parallel-job", "condition=Complete", "60s")

	log.Command("kubectl get job parallel-job -n playground")
	out, _ = kubectl.Get("job", "parallel-job", "-n", ns)
	log.Output(out)

	log.Command("kubectl get pods -n playground -l job-name=parallel-job")
	out, _ = kubectl.Get("pods", "-n", ns, "-l", "job-name=parallel-job")
	log.Output(out)
	log.Info("5 pods ran, but only 2 at a time. Total time ~3x faster than sequential.")
	prompt.StepPause()

	// ── Step 3: Failed Job with BackoffLimit ──────────────────────
	log.Section("Step 3: Failed Job with BackoffLimit")
	log.Concept(
		"Jobs that fail are retried with exponential backoff. The backoffLimit\n" +
			"controls how many retries before giving up. After exceeding backoffLimit,\n" +
			"the Job status becomes Failed.",
	)

	log.Step("Creating a Job that always fails (backoffLimit=4)...")
	if err := kubectl.Apply(filepath.Join(manifestDir, "03-job-backoff.yaml")); err != nil {
		log.Error("Failed: %v", err)
		os.Exit(1)
	}

	time.Sleep(3 * time.Second) // Give it time to fail a few times

	log.Command("kubectl get job backoff-job -n playground")
	out, _ = kubectl.Get("job", "backoff-job", "-n", ns)
	log.Output(out)

	log.Command("kubectl get pods -n playground -l job-name=backoff-job")
	out, _ = kubectl.Get("pods", "-n", ns, "-l", "job-name=backoff-job")
	log.Output(out)
	log.Info("Multiple pods = retries. Each retry waits longer (exponential backoff).")
	prompt.StepPause()

	// ── Step 4: CronJob ───────────────────────────────────────────
	log.Section("Step 4: CronJob — Scheduled Jobs")
	log.Concept(
		"A CronJob creates Jobs on a schedule (cron syntax). It's like Linux cron\n" +
			"but for Kubernetes workloads. The schedule '* * * * *' means every minute.",
	)

	log.Step("Creating CronJob (every minute)...")
	if err := kubectl.Apply(filepath.Join(manifestDir, "04-cronjob.yaml")); err != nil {
		log.Error("Failed: %v", err)
		os.Exit(1)
	}

	log.Command("kubectl get cronjob -n playground")
	out, _ = kubectl.Get("cronjob", "-n", ns)
	log.Output(out)

	log.Step("Waiting for at least one Job to be created (up to 60s)...")
	// CronJob fires at the start of each minute, so wait up to 60 seconds
	for i := 0; i < 6; i++ {
		out, _ = kubectl.Get("jobs", "-n", ns, "-l", "job-name")
		if strings.TrimSpace(out) != "" {
			break
		}
		time.Sleep(10 * time.Second)
	}

	log.Command("kubectl get jobs -n playground")
	out, _ = kubectl.Get("jobs", "-n", ns)
	log.Output(out)
	log.Info("The CronJob created a Job. Check back in a minute for the next one.")
	prompt.StepPause()

	// ── Step 5: TTL ───────────────────────────────────────────────
	log.Section("Step 5: TTL — Auto Cleanup")
	log.Concept(
		"ttlSecondsAfterFinished auto-deletes completed Jobs after a delay.\n" +
			"This prevents completed Job pods from accumulating and wasting resources.\n" +
			"All our Jobs have TTLs set for automatic cleanup.",
	)

	log.Command("kubectl get jobs -n playground")
	out, _ = kubectl.Get("jobs", "-n", ns)
	log.Output(out)
	log.Info("After their TTL expires, these Jobs will be automatically deleted.")
	prompt.StepPause()

	// ── Cleanup ───────────────────────────────────────────────────
	log.Section("Cleanup")
	log.Step("Deleting Jobs and CronJob...")
	kubectl.DeleteResource("job", "hello-job", ns)
	kubectl.DeleteResource("job", "parallel-job", ns)
	kubectl.DeleteResource("job", "backoff-job", ns)
	kubectl.DeleteResource("cronjob", "every-minute", ns)
	log.Success("All Jobs and CronJobs cleaned up!")

	// ── Summary ───────────────────────────────────────────────────
	log.Section("Summary: What You Learned")
	log.Info("  1. Jobs run tasks to completion (finite) vs Deployments (infinite)")
	log.Info("  2. completions = how many; parallelism = how many at once")
	log.Info("  3. backoffLimit controls retries with exponential backoff")
	log.Info("  4. CronJobs create Jobs on a schedule (cron syntax)")
	log.Info("  5. ttlSecondsAfterFinished auto-cleans up completed jobs")
	log.Info("  6. restartPolicy MUST be Never or OnFailure (not Always)")

	fmt.Println()
	log.Success("── Exercise 08 complete! ──")
	log.Info("Next: Module 09 — Probes & Resources")
	log.Info("  Run: go run ./exercises/09-probes-resources/")
	fmt.Println()
}
