# Architecture Decision Records

## ADR-001: ArgoCD over Flux

**Decision:** Use ArgoCD for GitOps.

**Reasoning:** ArgoCD has a superior UI for demos and portfolio presentations, making the GitOps sync loop visible at a glance. The App-of-Apps pattern is well-documented and widely used. ArgoCD is a CNCF graduated project with broader community adoption than Flux as of 2025.

**Trade-off:** Flux is equally valid in production and has better multi-tenancy support. Either would work here.

---

## ADR-002: Helm over raw Kustomize

**Decision:** Package the application as a Helm chart.

**Reasoning:** The `values.yaml` file doubles as the GitOps trigger — the CI pipeline writes a new `image.tag` to it, ArgoCD detects the diff and syncs. Helm also handles the PostgreSQL subchart dependency cleanly through `Chart.yaml` dependencies.

**Trade-off:** Kustomize patches are simpler for environments that don't need templating. For a single application, Kustomize + image overlays would also work.

---

## ADR-003: kind over minikube

**Decision:** Use kind (Kubernetes in Docker) for local cluster.

**Reasoning:** kind supports multi-node clusters (1 control-plane + 2 workers), starts faster than minikube, runs well in CI environments, and more accurately reflects a real multi-node cluster topology for a portfolio demo.

**Trade-off:** minikube has a built-in dashboard and easier ingress setup via addons. kind requires more manual configuration for ingress (not used in this project — port-forward is used instead).

---

## ADR-004: Monitoring as manual bootstrap, not ArgoCD-managed

**Decision:** Install kube-prometheus-stack via Helm directly in `scripts/setup.sh`, not as an ArgoCD-managed Application.

**Reasoning:** ArgoCD must not manage resources it did not create. If monitoring is installed manually first and then ArgoCD also has a monitoring Application, both become owners of the same Helm release — causing drift detection errors and reconciliation conflicts. Installing manually before ArgoCD bootstraps ensures:
1. The `ServiceMonitor` CRD exists before ArgoCD syncs the taskapi Application (which creates a ServiceMonitor resource)
2. Single clear ownership: kube-prometheus-stack is owned by Helm, taskapi is owned by ArgoCD

**Trade-off:** Monitoring state is not tracked in Git as a declarative resource. In production, you would use a separate GitOps tool or bootstrap operator for cluster-level addons, keeping application and platform concerns separate.
