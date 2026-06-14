# Module 12: Helm

## What Is Helm?

**Helm** is the package manager for Kubernetes. It bundles related Kubernetes manifests into a **chart** — a reusable, configurable package. Think `apt` or `brew` for Kubernetes.

```
helm install my-app ./myapp-chart --set replicaCount=5
```

## Core Concepts

### 1. Chart Structure

```
myapp-chart/
├── Chart.yaml           # Chart metadata (name, version, description)
├── values.yaml          # Default configuration values
├── values-prod.yaml     # Production override values
├── templates/           # Go-templated Kubernetes manifests
│   ├── _helpers.tpl    # Reusable template partials (naming, labels)
│   ├── deployment.yaml
│   ├── service.yaml
│   ├── configmap.yaml
│   ├── ingress.yaml
│   ├── hpa.yaml
│   └── NOTES.txt       # Post-install message
├── .helmignore          # Files to exclude from the chart package
```

### 2. Go Templating

Helm uses Go's `text/template` engine. Key elements:

```yaml
# Variable access
replicas: {{ .Values.replicaCount }}

# Conditionals
{{- if .Values.ingress.enabled }}
...
{{- end }}

# Loops
{{- range .Values.env }}
- name: {{ .name }}
  value: {{ .value }}
{{- end }}

# Template partials
{{ include "myapp.fullname" . }}

# Pipeline functions
{{ .Values.image.tag | default .Chart.AppVersion }}
```

The `-` before `{{` trims preceding whitespace.

### 3. values.yaml

```yaml
replicaCount: 3
image:
  repository: nginx
  tag: alpine
service:
  type: ClusterIP
  port: 80
ingress:
  enabled: false
```

Users override these with `--set` or `-f values-prod.yaml`:

```bash
helm install myapp ./myapp-chart --set replicaCount=5
helm install myapp ./myapp-chart -f values-prod.yaml
```

### 4. Release Management

| Command | Action |
|---------|--------|
| `helm install <name> <chart>` | Create a new release |
| `helm upgrade <name> <chart>` | Update an existing release |
| `helm rollback <name> <rev>` | Revert to a previous revision |
| `helm uninstall <name>` | Delete the release |
| `helm list` | List all releases |
| `helm history <name>` | Show release revision history |
| `helm lint <chart>` | Validate chart syntax |
| `helm template <name> <chart>` | Render templates locally (dry run) |
| `helm package <chart>` | Create a .tgz archive of the chart |

### 5. _helpers.tpl

Defines reusable named templates:

```yaml
{{- define "myapp.fullname" -}}
{{- .Release.Name }}-{{ .Chart.Name }}
{{- end -}}
```

Used in templates: `{{ include "myapp.fullname" . }}`

## What You'll Practice

1. Exploring the chart structure and understanding each file
2. Running `helm lint` to validate the chart
3. Running `helm template --dry-run` to see rendered YAML
4. Installing and upgrading a Helm release
5. Viewing release history and rolling back

## Key Gotchas

- **Helm 3 removed Tiller** — no server-side component. Releases are stored as Secrets in the cluster.
- **.Release.Name is immutable** — you can't rename a release after installation.
- **Whitespace matters in templates** — use `{{-` to trim it. Missing whitespace control causes malformed YAML.
- **values.yaml is the default, NOT the override** — `-f` and `--set` override values.yaml, not the other way around.
- **Helm doesn't manage CRDs automatically** — CRDs from the `crds/` directory are installed but never updated or deleted.
