# Agents Guide for a9s CLI

## Project in One Glance
- Go-based Cobra CLI (`a9s`) that provisions local/remote Kubernetes clusters, installs the a8s stack (cert-manager, Minio, PostgreSQL operator), and bootstraps Klutch control-plane demos on AWS.
- Goal: behave like a helpful assistant—automate multi-step workflows, surface generated assets, and stay usable across product releases without tight coupling.
- a9s Hub vision: self-service developer stacks and data services (PostgreSQL et al.) with cross-cluster provisioning via Klutch control planes.
- Primary user flows: `a9s create cluster a8s` (kind/minikube/EKS), `a9s create stack a8s` (install on existing cluster), PostgreSQL instance/backup/restore commands, and Klutch deploy/bind/delete for multi-cluster scenarios.

## Code Map
- `main.go`: entry; honors `DEBUG=1` for verbose logging.
- `cmd/`: Cobra commands and flag wiring (`root.go`, `create.go`, `pg.go`, `klutch.go`, etc.).
- `demo/`: orchestration layer (legacy naming) with `A8sDemoManager`; handles a8s installs, config management (`~/.a9s`), and `CheckPrerequisites()`.
- `creator/`: cluster provisioning abstractions (kind, minikube, EKS).
- `k8s/`: kubectl/client-go helpers and readiness polling; prefer `WaitFor*` helpers over sleeps.
- `pg/`, `minio/`: data-service resource generation/apply routines.
- `klutch/`: control-plane bootstrap (Crossplane, Dex, ingress) and cleanup/bind flows.
- `prerequisites/`: required tool detection with OS-specific install hints.
- `makeup/`: terminal UI utilities.
- Docs: `docs/` (Docusaurus, versioned); `ImplementationNotes.md` (historic, may drift); `Backlog.md` (open work); `.github/copilot-instructions.md` (expanded context).

## State and Configuration
- `demo.EstablishConfig()` persists `~/.a9s` with working dir and namespace; working dir holds cloned manifests (e.g., `a8s-deployment`).
- Global `UnattendedMode` (`-y/--yes`) skips prompts; still call `makeup.WaitForUser` before destructive actions.

## Working Norms for Agents
- Favor existing CLIs over reimplementation; keep outputs transparent (commands shown, configs accessible/editable).
- Use context-aware operations: pass kube contexts/namespaces explicitly; avoid assuming defaults.
- AWS resources must be tagged and use least-privilege IAM; reuse existing tagging patterns.
- Prefer readiness polling helpers (`WaitForKubernetesResource`, etc.) to fixed sleeps.
- Fatal paths should use `makeup.ExitDueToFatalError` to print and exit consistently.
- Treat `ImplementationNotes.md` as inspiration, not ground truth.

## Build, Test, Release
- Build: `make build` (injects `VERSION`, timestamp, last commit via ldflags); cross-compile with `make build_all`.
- Tests: `make test` or `make test_failfast`; coverage is sparse. Ruby e2e suite in `e2e-tests/` (run after `a9s create cluster`). Guard slow/destructive checks with tags/env toggles when adding new tests; prefer polling over sleeps.
- Docs site: `cd docs && yarn && yarn start/build`.
- Release: bump `VERSION`, update Readme/Changelog/docs, then tag `vX.Y.Z` on `main` to trigger GoReleaser GitHub action.

## Prerequisites & External Tools
- Baseline: `kubectl`, `helm`, `git`, plus `kind` or `minikube` depending on provider.
- Klutch/AWS: `kubectl-bind` (v1.4.1), `aws`, `eksctl`.
- Minio/S3 backups: ensure access to Minio or S3 endpoints; Minio is the default local option.
- Networked assets: `public.ecr.aws/w5n9a2g2/anynines` images, `https://charts.crossplane.io`, `https://github.com/anynines/a8s-deployment.git`.
- Key Go deps: `spf13/cobra`, `k8s.io/client-go`, `charmbracelet/lipgloss`, `goccy/go-yaml`.

## Pitfalls and Gotchas
- Context/namespace confusion is common; ensure flags pipe through to k8s helpers (system namespace `a8s-system`, instances default to `default`).
- Windows binaries exist but are largely untested; primary support is macOS/Linux.
- Backup/restore flows rely on object storage; Minio defaults require `path_style: true`; failed deletes/restores should surface clearly.
- Config regeneration: when backup-store params are provided, reconcile with existing `~/.a9s` config instead of silently reusing stale values.

## When in Doubt
- Check the backlog before inventing new work; align with design ideals (assistant-like, automation-first, minimal owned code).
- Keep user-visible artifacts (YAML manifests, commands) easy to inspect and override.
- Document new flags/flows in `Readme.md` and docs if behavior changes materially.
