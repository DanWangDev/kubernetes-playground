# 模块 09：探针与资源管理

## 什么是探针？

**探针**是 Kubernetes 对容器运行的健康检查。它们确定容器是否存活、能否服务流量以及是否已成功启动。

## 核心概念

### 1. 三种探针类型

| 探针 | 问题 | 失败动作 |
|------|------|----------|
| **Liveness** | 应用活着吗？（无死锁） | 重启容器 |
| **Readiness** | 应用能服务请求吗？ | 从 Service endpoints 移除 |
| **Startup** | 应用已完成启动？ | 成功前禁用 liveness/readiness |

### 2. 探针机制

```yaml
livenessProbe:
  httpGet:
    path: /healthz
    port: 8080
  initialDelaySeconds: 10
  periodSeconds: 5
```

三种机制：httpGet、exec（命令）、tcpSocket。

### 3. 探针时序

| 参数 | 含义 | 默认值 |
|------|------|--------|
| `initialDelaySeconds` | 首次探针前等待 | 0 |
| `periodSeconds` | 探针频率 | 10 |
| `failureThreshold` | 标记失败所需连续失败次数 | 3 |

### 4. 资源请求与限制

```yaml
resources:
  requests:        # 最低保障（调度器使用）
    cpu: 100m
    memory: 128Mi
  limits:          # 最大允许（超出则被限制/杀死）
    cpu: 500m
    memory: 256Mi
```

### 5. QoS 类别

| 类别 | 条件 | 驱逐优先级 |
|------|------|------------|
| **Guaranteed** | requests = limits | 最后（最安全） |
| **Burstable** | requests < limits | 中等 |
| **BestEffort** | 无 requests 或 limits | 最先被驱逐 |

## 你将练习的内容

1. 创建带就绪探针的 Pod
2. 观察失败的存活探针如何重启容器
3. 为慢启动应用使用启动探针
4. 设置资源请求和限制
5. 理解 QoS 类别

## 关键注意事项

- **不需要存活探针通常也没问题**——应用崩溃时进程退出，Kubernetes 会检测到
- **存活探针失败 = 重启**——设置保守的阈值
- **就绪探针失败 ≠ 重启**——Pod 保持运行但不接收流量
- **CPU 可压缩**（限制），**内存不可压缩**（OOMKilled）
