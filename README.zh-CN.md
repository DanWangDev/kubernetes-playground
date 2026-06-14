# Kubernetes 学习游乐场

通过指导式练习、IaC 清单和渐进式模块，从第一个 Pod 到生产模式，动手学习 Kubernetes。

## 快速开始

### 前提条件

- **Go 1.22+** ([下载](https://go.dev/dl/))
- **Docker** ([Docker Desktop](https://www.docker.com/products/docker-desktop/))
- **kubectl** ([安装指南](https://kubernetes.io/docs/tasks/tools/))

### 设置

```bash
git clone https://github.com/danwa/kubernetes-playground.git
cd kubernetes-playground
make cluster-create    # 创建本地 Kubernetes 集群（30 秒）
make cluster-status    # 验证集群就绪
```

### 运行第一个练习

```bash
make exercise-01                              # 非交互模式
go run ./exercises/01-first-pods/ --step      # 交互式逐步模式
```

## 学习路径

| 模块 | 主题 | 学习内容 |
|------|------|----------|
| 01 | 第一个 Pod | Pod 生命周期、kubectl 基础、标签、多容器 Pod |
| 02 | Deployments | 滚动更新、回滚、ReplicaSet、声明式管理 |
| 03 | Services | ClusterIP、NodePort、DNS、端口转发 |
| 04 | ConfigMap 与 Secret | 环境变量注入、卷挂载、不可变配置 |
| 05 | Ingress | L7 路由、NGINX 控制器、路径/主机路由、TLS |
| 06 | 存储 | PV、PVC、StorageClass、静态与动态供应 |
| 07 | StatefulSet | 稳定标识、有序部署/扩缩、每 Pod 存储 |
| 08 | Job 与 CronJob | 完成数、并行度、定时调度、TTL 清理 |
| 09 | 探针与资源 | 存活/就绪/启动探针、QoS 类别、OOMKilled |
| 10 | 自动扩缩 (HPA) | Metrics Server、CPU/内存目标、扩缩策略 |
| 11 | RBAC 与安全 | ServiceAccount、Role、NetworkPolicy、SecurityContext |
| 12 | Helm | Chart 结构、模板化、安装/升级/回滚 |
| 13 | 可观测性 | Prometheus、Grafana、ServiceMonitor、日志 Sidecar |
| 14 | 生产模式 | 亲和性、反亲和性、PDB、拓扑分散、污点容忍 |

## 项目结构

```
kubernetes-playground/
├── cmd/                  # CLI 工具（集群创建/删除/状态）
├── pkg/                  # 共享 Go 包
│   ├── kubectl/          # kubectl CLI 封装
│   ├── logger/           # 结构化练习输出
│   ├── prompt/           # 交互式分步模式
│   ├── validate/         # 断言和轮询辅助
│   └── env/              # 配置
├── exercises/            # 每个模块一个目录
│   ├── 01-first-pods/    # concept.md + concept_cn.md + manifests/ + main.go
│   ├── 02-deployments/
│   └── ...
├── test/                 # Go 测试（镜像 exercises/ 结构）
└── Makefile              # 任务运行器
```

## 可用命令

| 命令 | 描述 |
|------|------|
| `make cluster-create` | 创建 kind Kubernetes 集群 |
| `make cluster-delete` | 销毁 kind 集群 |
| `make cluster-reset` | 删除 + 重建（全新开始） |
| `make cluster-status` | 显示集群信息和节点 |
| `make exercise-NN` | 非交互式运行模块 |
| `go run ./exercises/NN-xxx/ --step` | 逐步交互运行 |
| `make test` | 运行所有测试 |
| `make lint` | 运行 go vet |
| `make build` | 构建所有包 |

## 设计决策

- **Go** — Kubernetes 的原生语言（kubectl、client-go、Helm 均为 Go）
- **kind** — Docker 中的 Kubernetes，无需虚拟机，CI 快速
- **原生 kubectl** — 练习通过真正的 kubectl 执行（教授 CLI，而非 SDK）
- **原始 YAML → Helm** — 模块 1-11 使用原始清单；模块 12 引入 Helm
- **渐进式模块** — 每个模块建立在前一个模块的概念之上
- **双语文档** — 每个模块均有 concept.md（英文）+ concept_cn.md（中文）

## 许可证

MIT
