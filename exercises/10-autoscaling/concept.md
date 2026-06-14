# Module 10: Autoscaling (HPA)

## What Is HPA?

The **Horizontal Pod Autoscaler (HPA)** automatically scales the number of pods in a Deployment or StatefulSet based on observed metrics (CPU, memory, or custom metrics).

```
HPA watches metrics → calculates desired replicas → scales the Deployment
```

## Core Concepts

### 1. HPA Algorithm

```
desiredReplicas = ceil(currentReplicas × currentMetricValue / desiredMetricValue)
```

Example: 1 replica using 80% CPU with a target of 50%:
```
desired = ceil(1 × 80 / 50) = ceil(1.6) = 2 replicas
```

### 2. HPA YAML

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: app-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: app-deployment
  minReplicas: 1
  maxReplicas: 5
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 50
```

### 3. Requirements

HPA needs:
1. **Metrics Server** installed in the cluster
2. **Resource requests** set on the target pods (HPA needs a baseline)

### 4. Metrics Server (kind)

```bash
kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml
```

Then verify:
```bash
kubectl top nodes
kubectl top pods
```

### 5. Stabilization and Scale-Down

HPA includes safeguards to prevent flapping:

```yaml
behavior:
  scaleDown:
    stabilizationWindowSeconds: 300  # Wait 5 min before scaling down
    policies:
    - type: Percent
      value: 50
      periodSeconds: 60
```

This means: wait 5 minutes after the last scale-up, then scale down by at most 50% per minute.

### 6. HPA Status

```
kubectl get hpa app-hpa
```

Shows current vs target utilization and current replica count. The HPA updates these every 15 seconds (default).

## What You'll Practice

1. (Prerequisite) Installing Metrics Server on kind
2. Creating a Deployment with CPU resource requests
3. Creating an HPA targeting 50% CPU
4. Viewing HPA status and understanding the metrics

## Key Gotchas

- **Metrics Server must be installed** — it's not included by default.
- **Pods need resource requests** — without requests, HPA has no baseline for percentage calculation.
- **HPA takes time** — metrics are collected every 15s, and stabilization windows add delay.
- **Scale down is slow by design** — prevents flapping. Scale up is fast.
- **HPA doesn't work with DaemonSets** — use VPA or manual scaling for those.
