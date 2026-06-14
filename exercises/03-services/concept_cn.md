# 模块 03：Service 与网络

## 什么是 Service？

Pod IP 是**短暂的**。当 Pod 死亡并被新 Pod 替换时，IP 会改变。**Service** 提供一个稳定的虚拟 IP 和 DNS 名称，通过标签选择器在一组 Pod 之间负载均衡流量。

把 Service 想象成一个拥有固定地址的负载均衡器。客户端连接到 Service，Service 将流量转发到健康的 Pod——无论任何时刻存在哪些具体的 Pod。

```
客户端 → Service（ClusterIP: 10.96.0.42）→ Pod A（10.244.1.5）
                                          → Pod B（10.244.1.6）
                                          → Pod C（10.244.2.3）
```

## 核心概念

### 1. Service 类型

| 类型 | 范围 | 用例 |
|------|------|------|
| **ClusterIP**（默认） | 仅集群内部 | 微服务间的后端通信 |
| **NodePort** | 外部通过 `<NodeIP>:<NodePort>` 访问 | 开发、演示、简单外部访问 |
| **LoadBalancer** | 外部通过云负载均衡器 | 生产流量入口（AWS NLB、GCP LB） |
| **ExternalName** | DNS CNAME | 指向外部服务 |

### 2. ClusterIP — 内部负载均衡

默认且最常见的类型。创建一个只能在集群内部访问的虚拟 IP：

```yaml
apiVersion: v1
kind: Service
metadata:
  name: nginx-svc
spec:
  type: ClusterIP
  selector:
    app: nginx           # 路由到具有此标签的 Pod
  ports:
  - port: 80             # Service 端口
    targetPort: 80       # Pod 上的容器端口
```

这将创建一个稳定的 DNS 名称：`nginx-svc.playground.svc.cluster.local`

### 3. DNS 解析

每个 Service 都获得一个 DNS A 记录。从集群中的任何 Pod：

```
# 全名（可从任何命名空间访问）
curl nginx-svc.playground.svc.cluster.local

# 短名称（可从相同命名空间访问）
curl nginx-svc
```

DNS 格式为：`<service>.<namespace>.svc.cluster.local`

### 4. Endpoints

Service 不直接将流量发送到 Pod。它们维护一个 **Endpoints**（或 **EndpointSlice**）对象，列出匹配 Pod 的 IP：

```
kubectl get endpoints nginx-svc
```

当 Pod 被添加或移除（Deployment 扩缩容），Endpoints 会自动更新。每个节点上的 kube-proxy 监视 Endpoints 并编写 iptables 规则来路由流量。

### 5. NodePort — 直接节点访问

NodePort 建立在 ClusterIP 之上，并在**每个节点**上额外开放一个端口（30000-32767）。到达 `<任意节点 IP>:<NodePort>` 的流量会到达 Service：

```yaml
spec:
  type: NodePort
  ports:
  - port: 80
    targetPort: 80
    nodePort: 30080      # 可选；如果省略则自动分配
```

NodePort 主要用于开发和演示。生产环境请使用 Ingress 或 LoadBalancer。

### 6. 端口转发

`kubectl port-forward` 创建一个从本地机器到 Pod 或 Service 的临时隧道：

```
kubectl port-forward svc/nginx-svc 8080:80 -n playground
```

这是一个开发工具——不用于生产。非常适合测试和调试。

### 7. 无头 Service（Headless Service）

带有 `clusterIP: None` 的 Service 不会获得虚拟 IP。DNS 直接解析为 Pod IP：

```yaml
spec:
  clusterIP: None
  selector:
    app: cassandra
```

无头 Service 是 StatefulSet 的构建块（Module 07）——每个 Pod 都有自己的 DNS 名称。

### 8. 多端口 Service

一个 Service 可以暴露多个具名端口：

```yaml
spec:
  ports:
  - name: http
    port: 80
    targetPort: 8080
  - name: metrics
    port: 9090
    targetPort: 9090
```

具名端口使意图清晰，并允许在不更改容器的情况下重新映射端口。

## 你将练习的内容

1. 创建 ClusterIP Service 并从集群内部访问它
2. 使用调试 Pod 测试内部 DNS 和连接性
3. 创建 NodePort Service 并从宿主机访问它
4. 使用 `kubectl port-forward` 进行本地开发访问
5. 理解 Endpoints 及其如何跟踪 Pod IP
6. 测试多端口 Service

## 关键注意事项

- **Service 选择器使用标签而非名称**——Service 通过匹配标签找到 Pod。如果标签不匹配，Service 没有 Endpoints，流量会静默失败。
- **NodePort 范围是 30000-32767**——不能使用 80 端口作为 NodePort。kind 会自动将 NodePort 映射到宿主机。
- **ClusterIP 不是真实的网络接口**——它是由 kube-proxy 通过 iptables/IPVS 实现的虚拟 IP。可以 ping Pod IP，但不能 ping ClusterIP。
- **port vs targetPort**——`port` 是 Service 的监听端口；`targetPort` 是容器的端口。它们可以不同（如 Service 端口 80 → 容器端口 8080）。
- **应用中的 DNS 缓存**——某些应用缓存 DNS 查找，不会看到更新的 Endpoints。对于这些情况，使用 `targetPort` 和无头 Service。
- **EndpointSlice vs Endpoints**——Kubernetes 1.21+ 使用 EndpointSlices（更小的对象，更好的伸缩性）。Endpoints 仍然保留用于向后兼容。
