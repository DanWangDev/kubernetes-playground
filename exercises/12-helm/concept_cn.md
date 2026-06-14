# 模块 12：Helm

## 什么是 Helm？

**Helm** 是 Kubernetes 的包管理器。它将相关的 Kubernetes 清单打包成 **chart**——可复用、可配置的包。相当于 Kubernetes 的 `apt` 或 `brew`。

## 核心概念

### 1. Chart 结构

```
myapp-chart/
├── Chart.yaml           # Chart 元数据
├── values.yaml          # 默认配置值
├── templates/           # Go 模板化的 Kubernetes 清单
│   ├── _helpers.tpl    # 可复用模板片段
│   ├── deployment.yaml
│   ├── service.yaml
│   └── NOTES.txt       # 安装后消息
└── .helmignore
```

### 2. Go 模板

```yaml
replicas: {{ .Values.replicaCount }}
{{- if .Values.ingress.enabled }}
...
{{- end }}
```

### 3. values.yaml

```yaml
replicaCount: 3
image:
  repository: nginx
  tag: alpine
```

通过 `--set` 或 `-f` 覆盖：

```bash
helm install myapp ./myapp-chart --set replicaCount=5
```

### 4. 发布管理

| 命令 | 操作 |
|------|------|
| `helm install` | 创建新发布 |
| `helm upgrade` | 更新已有发布 |
| `helm rollback` | 回滚到之前版本 |
| `helm uninstall` | 删除发布 |
| `helm list` | 列出所有发布 |
| `helm history` | 显示发布版本历史 |

## 你将练习的内容

1. 探索 chart 结构并理解每个文件
2. 运行 `helm lint` 验证 chart
3. 运行 `helm template` 查看渲染后的 YAML
4. 安装和升级 Helm 发布
5. 查看发布历史并回滚

## 关键注意事项

- **Helm 3 移除了 Tiller**——发布存储为集群中的 Secret
- **模板中空白很重要**——使用 `{{-` 修剪空白
- **values.yaml 是默认值，不是覆盖值**——`-f` 覆盖 values.yaml
