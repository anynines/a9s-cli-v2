# AGENTS Guide

### Mental Model (what each piece does)
- `a9s-cli-v2`: Go Cobra CLI orchestrating clusters (kind/minikube/EKS), a8s stack installs, and Klutch control-plane demos. Core packages: `cmd/` (flags + UX), `demo/` (state in `~/.a9s`), `creator/` (cluster provisioners), `k8s/` (kubectl/client-go helpers), `pg/` (Postgres flows), `klutch/` (control-plane bootstrap + binding), `makeup/` (TTY output).
- `backend-kube-bind` (anynines-backend): kube-bind compatible control-plane backend that sits behind an OIDC issuer. Local demos typically use Dex; the AWS flow fronts it with Cognito (no Dex in that path). Exposes `/bind` (interactive) and `/bind-noninteractive` (machine flow) and returns kubeconfigs for workload clusters. Lives inside the control-plane cluster, uses the `oidc-config` and `cookie-config` secrets, and currently enforces clientID on the interactive path.
- `a9s-tenants-operator`: Kubebuilder operators plus Tenant CRD. Reconciles Tenants into IdPs (Cognito/Keycloak), emits `status.oidc` and a Secret reference with client secret, and tags resources for the control plane. Must start only if it can list Tenants (RBAC/CRD check).

### System Architecture (how they fit)
- Control plane (local kind or AWS EKS) runs Crossplane + anynines config packages, anynines-backend, ingress (NGINX or ALB), and the tenant operator. OIDC issuer differs: Dex for local demo, Cognito for AWS (no Dex there). AWS flow uses IRSA, ALB, and can derive bind URLs from Route53 hosted zones.
- Workload binding:
  - Interactive: `a9s bind klutch workload --interactive-bind` shells out to `kubectl bind` (browser + Dex).
  - Non-interactive: `a9s bind klutch workload` uses tenant OIDC client credentials to call `/bind-noninteractive`, then deploys the kube-bind konnector and can persist the returned kubeconfig.
- State/config: `demo.EstablishConfig()` persists `~/.a9s` for workspace paths and defaults; all commands should honor `--context`, `--namespace`, and `--yes` (unattended).

### AWS Provider Deep Dive (Klutch)
- Control-plane EKS bill of materials: VPC with DNS support, 3 public + 3 private subnets, IGW, per-AZ NAT gateways + EIPs, public/private route tables, cluster SG, IAM roles (cluster, node, KMS inline, ALB controller policy, tenant-operator IRSA), KMS key for secret encryption, managed nodegroup, gp3 default StorageClass, ALB controller (OIDC provider association + helm), ALB ingress, kubeconfig update and node readiness. Deploy tenant operator (IRSA) and derive bind URL/request from hosted zone or ConfigMap.
- Workload EKS: same tagging (`Klutch=<role>`, `eks.cluster/name|id`), networking scaffold and node shape/count tuning; skips tenant-operator deployment.
- Addons path: on an existing EKS cluster, associate OIDC provider, ensure tenant-operator IAM role, derive bind URL/request (ConfigMap or hosted zone), deploy tenant-operator Helm chart (ECR login when chart is OCI).
- Cognito tenants: reuse/create tagged user pool (`Klutch=ControlPlane`), resource server scope `klutch/bind`, per-tenant app clients (client_credentials), hosted domain, Secrets Manager entry with issuer/tokenURL/clientID/secret/scope/bindURL/bindRequest (default secret `klutch/<tenant>/oidc-client`).
- Hosted zone/DNS: delegation check uses public resolvers (8.8.8.8/1.1.1.1), compares live NS to Route53, prints exact NS records to set in the parent, and polls up to 30m before failing; CNAME wait loops until target observed or timeout.
- Teardown: remove gp3 SC, ALB controller release + IAM SA/policy, tenant-operator IAM if present, managed nodegroups, EKS cluster, VPC artifacts (NAT/IGW/RT/subnets/SG), optional Route53 records/hosted zone/ACM cert, optional orphaned EIP cleanup. Dry-run prints the full plan without changes.

### OIDC/Tenant Model (cross-repo contract)
- Multi-tenant OIDC: one issuer per control plane (Dex locally, Cognito on AWS); per-tenant app clients + scopes (e.g., `klutch/bind:<tenant_uuid>`). Backend must accept multiple client IDs and Cognito access tokens (`token_use=access`) on `/bind-noninteractive` while keeping strict verifier for interactive `/bind`.
- Backend should source allowed client IDs/scopes from Tenant CR status (or interim static allowlist) and still enforce issuer + scope + client_id allowlist. See ADRs 0001/0002/0003/0005/0006.
- Tenant operator must tag IdP assets, be idempotent, and keep Tenant data cached/informers for backend consumption; fail fast if it cannot list Tenants.

### Development Constraints and Practices
- Assistant mindset: automate end-to-end; prefer existing CLIs (kubectl/helm/aws/eksctl/kube-bind) over reimplementation; show commands and generated manifests so users can inspect/override.
- Keep AWS changes least-privileged and tagged; reuse resources when present; honor dry-run paths; validate required CLIs before mutating anything.
- Avoid tight coupling to product releases: keep versions overrideable via flags/envs; only pin when required (ALB controller, tenant-operator chart/image).
- Use readiness polling helpers in `k8s/` and `makeup.ExitDueToFatalError` for fatal paths; avoid sleeps.
- Wire Cobra flags through to business logic and helpers; respect `--yes` (no prompts) but still require confirmation for destructive steps unless explicitly bypassed.
- Documentation: update `Readme.md`/`docs/` and CLI help when behavior changes; keep examples runnable.
- Build + tests required for every change: code must compile and tests must exist, stay current, and pass; if no relevant tests exist, add them or document the gap explicitly.
- Architecture decisions: create or update ADRs in `docs/adr/` when introducing non-trivial behavior changes, new contracts, or security/compatibility impacts.
- If the program flow depends on external artefacts such as container images, helm charts etc. implement preflight validation to verify their existence. Provide meaningful error messages if not and fail fast.

### Repo-Specific Guidance
- `a9s-cli-v2`: Go 1.22; build with `make build` (binary in `bin/a9s`). Tests are sparse (`make test`). Preserve logging/UX (`makeup`), config under `~/.a9s`, and kube context/namespace plumbing. Klutch AWS logic sits in `klutch/aws/`; binding UX in `cmd/bind_klutch_workload.go`.
- `backend-kube-bind`: Go HTTP service; interactive path stays strict on clientID. Add a lax verifier/allowlist for `/bind-noninteractive` that accepts Cognito access tokens, enforces issuer + scope (`klutch/bind*`) + client_id allowlist. Runs with cluster-admin via `klutch/templates/backend.tmpl`; secrets `oidc-config` and `cookie-config` must stay in sync with CLI/bootstrap.
- `a9s-tenants-operator`: Kubebuilder; on startup, list Tenants to ensure CRD/RBAC is present. Reconcile per-provider resources, tag them, write client secret to Secret, and populate `status.oidc`. Keep reconciliation idempotent and retry-safe; avoid per-request IdP calls by caching.

### Collaboration Habits for AI Contributions
- Start with `rg` and small, targeted diffs (`apply_patch`); keep changes ASCII and terse. Comment only when logic is non-obvious.
- For AWS paths, never skip dry-run handling; gate destructive actions behind confirmations unless `--yes` is set.
- Add focused tests where practical; for destructive flows, guard with env flags or dry-run switches. Keep new flags/plumbing consistent across Cobra, execution, and docs.
- Keep create/delete flows paired: anything created should have a corresponding delete path (unless explicitly intended to remain, e.g., DNS/ACM when opted out). When reusing resources (IAM roles, OIDC providers, hosted zones), reconcile or validate them for consistency with the current cluster; if validation is expensive or brittle, prefer a clean delete/recreate path and document the choice.
- Do not "fix" a failing primary path by inventing alternate execution paths (for example, retrying with a local chart when the OCI chart is expected). Stick to the intended flow, explain why it is failing, and list concrete options to fix it (e.g., mispublished artifacts, wrong version/tag, missing auth/permissions, or required preflight checks).

### Handy Commands
- Build: `make build` (binary at `bin/a9s`); cross-compile with `make build-all`.
- Tests: `make test` or `make test_failfast`; e2e (Ruby) in `e2e-tests` after `a9s create cluster`.
- Docs: `cd docs && yarn && yarn start/build`.
