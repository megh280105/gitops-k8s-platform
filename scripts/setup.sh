#!/usr/bin/env bash
set -euo pipefail

RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'; NC='\033[0m'
info()    { echo -e "${GREEN}[INFO]${NC} $*"; }
warn()    { echo -e "${YELLOW}[WARN]${NC} $*"; }
die()     { echo -e "${RED}[ERROR]${NC} $*" >&2; exit 1; }

# ── Prerequisites check ────────────────────────────────────────────────────────
info "Checking prerequisites..."
for cmd in docker kind kubectl helm; do
  command -v "$cmd" &>/dev/null || die "$cmd is not installed. Run: brew install $cmd"
done
docker info &>/dev/null || die "Docker is not running. Start Docker Desktop first."
info "All prerequisites OK"

CLUSTER_NAME="${CLUSTER_NAME:-gitops}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"

# ── Phase 4a: Create kind cluster ─────────────────────────────────────────────
if kind get clusters 2>/dev/null | grep -q "^${CLUSTER_NAME}$"; then
  warn "kind cluster '${CLUSTER_NAME}' already exists, skipping creation"
else
  info "Creating kind cluster '${CLUSTER_NAME}'..."
  kind create cluster --config "$ROOT_DIR/kind-config.yaml" --name "$CLUSTER_NAME"
fi

kubectl cluster-info --context "kind-${CLUSTER_NAME}" &>/dev/null || die "Cannot reach cluster"
info "Cluster ready: $(kubectl get nodes --no-headers | wc -l | tr -d ' ') nodes"

# ── Phase 4b: Install kube-prometheus-stack (manual — NOT via ArgoCD) ─────────
# Reason: ensures ServiceMonitor CRD exists before ArgoCD syncs taskapi
info "Adding Helm repos..."
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo add argo https://argoproj.github.io/argo-helm
helm repo add bitnami https://charts.bitnami.com/bitnami
helm repo update

if helm status kube-prometheus-stack -n monitoring &>/dev/null; then
  warn "kube-prometheus-stack already installed, skipping"
else
  info "Installing kube-prometheus-stack (this pulls several images, may take 5-10 min)..."
  helm install kube-prometheus-stack prometheus-community/kube-prometheus-stack \
    -n monitoring --create-namespace \
    -f "$ROOT_DIR/monitoring/values.yaml" \
    --timeout 15m --wait
fi

info "Waiting for ServiceMonitor CRD..."
kubectl wait crd/servicemonitors.monitoring.coreos.com \
  --for=condition=Established --timeout=120s

# ── Phase 5a: Install ArgoCD ───────────────────────────────────────────────────
if helm status argocd -n argocd &>/dev/null; then
  warn "ArgoCD already installed, skipping"
else
  info "Installing ArgoCD..."
  helm install argocd argo/argo-cd \
    -n argocd --create-namespace \
    --timeout 10m --wait
fi

info "Waiting for ArgoCD server..."
kubectl wait --for=condition=available deploy/argocd-server -n argocd --timeout=120s

# ── Phase 5b: Apply App-of-Apps ────────────────────────────────────────────────
info "Applying App-of-Apps..."
kubectl apply -f "$ROOT_DIR/argocd/apps/app-of-apps.yaml"

# ── Done ───────────────────────────────────────────────────────────────────────
ARGOCD_PASS=$(kubectl -n argocd get secret argocd-initial-admin-secret \
  -o jsonpath='{.data.password}' | base64 -d)

echo ""
echo -e "${GREEN}╔══════════════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║          Setup complete! Access your services:       ║${NC}"
echo -e "${GREEN}╚══════════════════════════════════════════════════════╝${NC}"
echo ""
echo "  ArgoCD UI:"
echo "    kubectl port-forward svc/argocd-server -n argocd 8080:443"
echo "    open https://localhost:8080"
echo "    user: admin | pass: ${ARGOCD_PASS}"
echo ""
echo "  Grafana:"
echo "    kubectl port-forward svc/kube-prometheus-stack-grafana -n monitoring 3000:80"
echo "    open http://localhost:3000  (admin / admin)"
echo ""
echo "  Task API (once ArgoCD syncs taskapi):"
echo "    kubectl port-forward svc/taskapi-taskapi -n taskapi 8888:80"
echo "    curl http://localhost:8888/health"
echo ""
