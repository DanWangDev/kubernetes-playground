# 模块 13：可观测性

## 什么是可观测性？

可观测性意味着从系统的外部输出理解系统内部发生的事。在 Kubernetes 中包括：

- **日志**：应用写入 stdout/stderr 的内容
- **指标**：随时间变化的数值测量（CPU、内存、请求率）
- **事件**：Kubernetes 集群级别的发生记录
- **追踪**：跨服务的分布式请求流

## 核心概念

### 1. kubectl logs — 第一道防线

```
kubectl logs <pod>                    # 当前容器日志
kubectl logs <pod> -c <container>     # 多容器 Pod 中的特定容器
kubectl logs <pod> --tail=50          # 最后 50 行
kubectl logs <pod> --since=5m         # 最近 5 分钟
kubectl logs <pod> --previous         # 前一个（崩溃的）容器的日志
```

### 2. kubectl get events — 集群活动日志

事件显示：调度决策、镜像拉取、探针结果、卷挂载、扩缩操作。这是排查问题的第一站。

### 3. kubectl describe — 深度检查

显示完整的 Pod 状态：条件、容器状态、最近事件、资源使用。

### 4. Prometheus + Grafana

标准 Kubernetes 监控栈：
- **Prometheus**：从 Pod 抓取指标并存储时序数据
- **Grafana**：在 Prometheus 数据之上构建仪表盘和告警

### 5. 结构化日志

将 JSON 写入 stdout。日志聚合工具（Elasticsearch、Loki、Datadog）能更好地解析结构化日志。

### 6. Sidecar 日志模式

当应用将日志写入文件（而非 stdout）时，使用 sidecar 容器将文件 tail 到 stdout。

## 你将练习的内容

1. 使用 kubectl logs 的各种标志
2. 将集群事件视为调试工具
3. 使用 kubectl describe 深度检查 Pod
4. 理解 Prometheus 架构
5. 部署日志 sidecar 模式

## 关键注意事项

- **事件不是持久的**——默认 1 小时后删除
- **`--previous` 非常宝贵**——容器崩溃重启后，当前日志来自新容器
- **Prometheus 资源密集**——全量安装需要大量资源
