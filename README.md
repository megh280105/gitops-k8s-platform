# GitOps Kubernetes Platform

A production-quality Task Manager API deployed to Kubernetes using a full GitOps loop — built as a DevOps portfolio project.

```
Push code
  └─▶ GitHub Actions: test → build → push image to ghcr.io → update Helm values
                                                                      │
                                                               ArgoCD detects diff
                                                                      │
                                                              New pods roll out
                                                                      │
                                                       Prometheus + Grafana monitor
```

## Stack

| Layer | Technology |
|---|---|
| App | Go 1.26 REST API (Tasks CRUD) |
| Container | Docker (multi-stage, non-root) |
| Registry | ghcr.io (free for public repos) |
| Kubernetes | kind (local, 3-node) |
| GitOps | ArgoCD — App-of-Apps pattern |
| Packaging | Helm chart + Bitnami PostgreSQL subchart |
| CI/CD | GitHub Actions |
| Monitoring | Prometheus + Grafana (kube-prometheus-stack) |
| IaC | Terraform — AWS VPC + EKS modules (portfolio artifact) |

## Quick Start

### Prerequisites
```bash
brew install go kind kubectl helm
# Docker Desktop must be running
```

### Run locally (no Kubernetes)
```bash
# Start PostgreSQL
docker run -d -e POSTGRES_DB=tasks -e POSTGRES_USER=taskapi \
  -e POSTGRES_PASSWORD=changeme -p 5432:5432 postgres:16-alpine

cd app
DATABASE_URL="postgres://taskapi:changeme@localhost:5432/tasks?sslmode=disable" go run .
curl localhost:8080/health
```

### Full GitOps setup on kind
```bash
./scripts/setup.sh
```

This creates a kind cluster, installs Prometheus + Grafana, installs ArgoCD, and applies the App-of-Apps manifest. ArgoCD then syncs the taskapi Helm chart automatically.

### Teardown
```bash
./scripts/teardown.sh
```

## Accessing services

After `setup.sh` completes:

```bash
# ArgoCD UI
kubectl port-forward svc/argocd-server -n argocd 8080:443
# https://localhost:8080  |  user: admin  |  pass: (printed by setup.sh)

# Grafana
kubectl port-forward svc/kube-prometheus-stack-grafana -n monitoring 3000:80
# http://localhost:3000  |  admin / admin

# Task API (once ArgoCD syncs)
kubectl port-forward svc/taskapi-taskapi -n taskapi 8888:80
curl localhost:8888/health
curl localhost:8888/tasks
```

## API Reference

| Method | Path | Description |
|---|---|---|
| GET | `/health` | Liveness probe |
| GET | `/ready` | Readiness probe (checks DB) |
| GET | `/metrics` | Prometheus metrics |
| GET | `/tasks` | List all tasks |
| POST | `/tasks` | Create task `{"title":"...","description":"..."}` |
| GET | `/tasks/{id}` | Get task by ID |
| PUT | `/tasks/{id}` | Update task |
| DELETE | `/tasks/{id}` | Delete task |

## GitOps Loop

1. Push a code change to `main`
2. GitHub Actions runs `go test`, builds Docker image, pushes to `ghcr.io`
3. CI commits updated `image.tag` in `helm/taskapi/values.yaml` with `[skip ci]`
4. ArgoCD detects the diff (polls every 3 min or via webhook) and syncs
5. Kubernetes rolls out new pods; old pods terminate after health checks pass

## Dev-only Assumptions

| Assumption | Production equivalent |
|---|---|
| PostgreSQL password `changeme` in values.yaml | External secret from Vault / AWS Secrets Manager |
| Grafana password `admin` in values.yaml | Sealed secret / external secrets operator |
| Terraform uses local backend | S3 + DynamoDB remote state |
| Terraform never applied | Real AWS account + IAM |

## Architecture Decisions

See [docs/adr.md](docs/adr.md) for decisions on: ArgoCD vs Flux, Helm vs Kustomize, kind vs minikube, and why monitoring is a manual bootstrap dependency.

## Terraform (Portfolio Artifact)

The `terraform/` directory contains production-quality AWS EKS modules (VPC + EKS 1.31) that are never applied in this local setup. They demonstrate IaC design patterns for a portfolio audience.

```bash
cd terraform/environments/dev
terraform init -backend=false
terraform validate   # passes without AWS credentials
```
