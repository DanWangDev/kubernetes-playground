# 模块 05：Ingress 与 HTTP 路由

## 什么是 Ingress？

**Ingress** 为 Service 提供 7 层（HTTP/HTTPS）路由。与 Service（L4）不同，Ingress 理解 HTTP——它可以按主机名、路径、请求头进行路由，并终结 TLS。

Ingress 包含两部分：
1. **Ingress Controller** — 执行实际路由的 Pod（nginx、traefik、haproxy 等）
2. **Ingress 资源** — 定义路由规则的 YAML

## 核心概念

### 1. L4 vs L7

| | Service（L4） | Ingress（L7） |
|---|---|---|
| **OSI 层** | 传输层（TCP/UDP） | 应用层（HTTP） |
| **路由方式** | 按 IP:port | 按主机名 + 路径 |
| **TLS** | 透传 | 终结 + 重新加密 |
| **功能** | 负载均衡 | 路由、TLS、限流、重写、认证 |

### 2. Ingress Controller 安装（kind）

Kubernetes 默认不包含 Ingress Controller。对于 kind：

```bash
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml
```

这将在 `ingress-nginx` 命名空间中部署 NGINX Ingress Controller，监听 kind 节点的 80 和 443 端口。

### 3. 基于路径的路由

```yaml
spec:
  rules:
  - http:
      paths:
      - path: /api
        pathType: Prefix
        backend:
          service:
            name: api-service
            port:
              number: 80
```

pathType 选项：
- **Prefix**：匹配路径前缀
- **Exact**：精确匹配路径
- **ImplementationSpecific**：控制器特定匹配

### 4. 基于主机的路由

```yaml
spec:
  rules:
  - host: api.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: api-service
            port:
              number: 80
```

### 5. TLS 终结

Ingress Controller 终结 TLS 并将明文 HTTP 转发到后端：

```yaml
spec:
  tls:
  - hosts:
    - app.example.com
    secretName: app-tls-secret
```

## 关键注意事项

- **Ingress Controller 不是内置的**——必须自行安装。
- **Ingress 和 Service 工作在不同层级**——Ingress 将 HTTP 路由到 Service；Service 将 TCP 路由到 Pod。
- **defaultBackend**——如果没有匹配的规则，默认后端处理请求。
- **TLS Secret 必须与 Ingress 在同一命名空间**。
- **路径顺序很重要**——`/` 匹配所有内容。将更具体的路径放在前面。
