# 模块 01：第一个 Pod

## 什么是 Pod？

在 Kubernetes 中，**Pod** 是最小的可部署单元。它是一个或多个容器的组合，共享：

- **网络命名空间** — 相同的 IP 地址和端口空间
- **IPC 命名空间** — 可通过 SystemV 信号量或 POSIX 共享内存通信
- **存储卷** — 共享的文件系统挂载

将 Pod 视为一个逻辑主机：它运行你的应用容器，就像虚拟机一样，但更轻量、更专注。大多数 Pod 运行单个容器，但当你需要紧密耦合（边车、代理、日志收集器）时，多容器 Pod 非常常见。

## 核心概念

### 1. Pod 生命周期

每个 Pod 都会经历定义好的生命周期：

```
Pending → Running → Succeeded（或 Failed）
```

| 阶段 | 含义 |
|------|------|
| **Pending** | Pod 已被 API 服务器接受，但一个或多个容器尚未运行。包括镜像拉取和调度时间。 |
| **Running** | Pod 已绑定到节点，所有容器已创建，至少有一个仍在运行。 |
| **Succeeded** | 所有容器成功终止（退出码为 0）。仅适用于 restartPolicy: Never 或 OnFailure。 |
| **Failed** | 所有容器终止，至少有一个非零退出码。 |
| **Unknown** | 无法确定 Pod 状态（通常是节点通信错误）。 |

查看 Pod 阶段：

```
kubectl get pod <name> -o jsonpath='{.status.phase}'
```

### 2. kubectl：你的日常工具

`kubectl` 是与 Kubernetes API 服务器通信的 CLI。以下是你最常用的命令：

| 命令 | 用途 |
|------|------|
| `kubectl get pods` | 列出当前命名空间中的 Pod |
| `kubectl get pods -n <ns>` | 列出特定命名空间中的 Pod |
| `kubectl get pods -o wide` | 显示更多细节（节点、IP） |
| `kubectl describe pod <name>` | Pod 完整信息，包括事件 |
| `kubectl logs <name>` | 获取第一个容器的 stdout |
| `kubectl logs <name> -c <container>` | 获取特定容器的 stdout |
| `kubectl exec <name> -- <command>` | 在 Pod 中运行命令 |
| `kubectl delete pod <name>` | 删除 Pod |
| `kubectl apply -f <file>` | 从 YAML 文件创建或更新资源 |

### 3. 命名空间（Namespaces）

命名空间将集群资源划分给多个用户或项目。它们就像文件夹：

- 不同命名空间中的资源彼此隔离
- 命名空间级资源包括 Pod、Deployment、Service、ConfigMap、Secret
- 集群级资源（Node、PV、Namespace 本身）不属于命名空间
- 默认命名空间是 `default`——使用命名空间来组织工作

```
kubectl get namespaces
kubectl create namespace playground
```

### 4. Pod YAML 结构

每个 Pod 清单都有四个必需的顶级字段：

```yaml
apiVersion: v1          # 此资源类型的 API 版本
kind: Pod               # 资源类型
metadata:               # 标识信息
  name: my-pod
  labels:
    app: demo
spec:                   # 期望状态
  containers:
  - name: nginx
    image: nginx:alpine
    ports:
    - containerPort: 80
```

### 5. 标签和选择器（Labels and Selectors）

**标签**是附加到 Kubernetes 对象上的键值对。它们是连接资源的粘合剂：

```yaml
labels:
  app: web
  environment: staging
  tier: frontend
```

标签实现了松耦合——Service 通过标签选择器找到目标 Pod，而不是通过 Pod 名称或 IP。你会经常使用它们：

```
kubectl get pods -l app=web
kubectl get pods -l 'environment in (staging, production)'
```

### 6. 多容器 Pod

一个 Pod 可以包含多个共享命运的容器——它们位于同一节点，共享同一网络命名空间（可通过 `localhost` 互相访问），并可以共享卷。常见模式：

- **边车（Sidecar）**：增强主容器（如日志收集器、代理、配置重载器）
- **大使（Ambassador）**：代理主容器的进出流量
- **适配器（Adapter）**：规范化主容器的输出

```yaml
spec:
  containers:
  - name: app
    image: my-app:latest
  - name: sidecar
    image: busybox
    command: ["tail", "-f", "/var/log/app.log"]
```

## 你将练习的内容

1. 创建命名空间来组织工作
2. 启动一个简单的 nginx Pod 并检查其生命周期
3. 为 Pod 添加标签并按标签选择器查询
4. 在运行中的容器内执行命令
5. 检查 Pod 日志和事件历史
6. 部署多容器 Pod（应用 + 边车模式）
7. 清理资源

## 关键注意事项

- **Pod 不可变**——创建后无法更改 Pod 的 spec。编辑 YAML，删除 Pod，然后重新创建。（Module 02 中的 Deployment 解决了这个问题。）
- **Pod 是短暂的**——当 Pod 死亡时，它永远消失了（包括其 IP、本地存储和日志，除非有集中式日志）。不要依赖 Pod IP。
- **Get vs Describe**——`kubectl get` 提供摘要。`kubectl describe` 提供完整信息，包括事件日志（调度决策、镜像拉取进度、探针结果）。
- **命名空间不会自动创建**——必须先创建命名空间才能使用它。应用引用不存在命名空间的 YAML 会失败。
- **镜像拉取需要时间**——第一次在节点上运行镜像时，Kubernetes 必须下载它。如果镜像不可用，状态会显示 `ErrImagePull` 或 `ImagePullBackOff`。
- **localhost 跨容器工作**——在多容器 Pod 中，容器通过 `localhost:<port>` 互相访问。这是因为它们共享网络命名空间。
