# 模块 08：Job 与 CronJob

## 什么是 Job？

**Job** 创建一个或多个 Pod，并确保指定数量成功完成。与 Deployment（保持 Pod 持续运行）不同，Job 用于有限任务：批处理、数据库迁移、报表生成。

**CronJob** 按计划创建 Job——类似于 cron，但是为 Kubernetes 设计的。

## 核心概念

### 1. Job 模式

| 模式 | completions | parallelism | 用例 |
|------|-------------|-------------|------|
| **单次** | 1 | 1 | 一次性任务（迁移、备份） |
| **固定计数** | N | M | 用 M 个 Worker 处理 N 个项目 |
| **工作队列** | —（不设） | M | 从队列处理项目直到为空 |

### 2. Job YAML

```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: hello-job
spec:
  completions: 1
  parallelism: 1
  backoffLimit: 4
  template:
    spec:
      containers:
      - name: hello
        image: busybox
        command: ["echo", "Hello, Kubernetes!"]
      restartPolicy: Never
```

### 3. Job 生命周期

Job 创建 → Pod 创建 → Pod 完成 → Job 检查 completions
→ 完成：Job = Complete
→ Pod 失败：重试（最多 backoffLimit 次）

### 4. CronJob

```yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: daily-report
spec:
  schedule: "0 6 * * *"
  concurrencyPolicy: Forbid
  successfulJobsHistoryLimit: 3
  jobTemplate:
    spec:
      template: ...
```

### 5. 并发策略

| 策略 | 行为 |
|------|------|
| **Allow**（默认） | 多个 Job 可并发运行 |
| **Forbid** | 如前一个仍在运行则跳过 |
| **Replace** | 终止前一个并启动新的 |

## 你将练习的内容

1. 创建简单的一次性 Job
2. 运行多 Worker 并行 Job
3. 观察失败的 Job 及其退避限制
4. 创建 CronJob 并观察其生成 Job
5. 配置 TTL 实现自动清理

## 关键注意事项

- **restartPolicy 必须是 Never 或 OnFailure**——`Always`（默认值）对 Job 无效
- **Job Pod 会保留**——已完成/失败的 Job Pod 保留以供日志检查。使用 TTL 自动清理
- **CronJob 历史会累积**——设置 `successfulJobsHistoryLimit` 和 `failedJobsHistoryLimit`
