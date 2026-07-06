# ADR 0001: Enforce Tenant CR Access and Prepare Backend Tenant Awareness

## Status
Accepted

## Context
- The tenant operator reconciles `Tenant` CRs to provision Cognito app clients and secrets.
- The upcoming tenant-aware anynines-backend will validate tokens per-tenant by reading `Tenant` CR data from the control-plane cluster.
- If RBAC prevents access to `Tenant` CRs, reconciliation would silently stall and the backend would fail to authorize workload clusters.
- We need a hard failure when CR access is missing, and a clear contract that the backend will rely on the same CRD for future authorization features without expensive per-request lookups.

## Decision
- On operator startup, perform a capability check: attempt to list `Tenant` resources in the operator’s namespace. If this fails (RBAC, missing CRD, or API error), exit immediately.
- Keep using the informer/cache for `Tenant` CRs so reconciliation and future backend lookups remain in-memory and low latency.
- Document that the backend will consume `Tenant` CR data for per-tenant token verification, enabling future tenant-based authorization without per-request cascades.

## Consequences
- Misconfigured RBAC or missing CRDs fail fast, making the problem obvious.
- The operator and future backend tenants will avoid per-request API calls by using cached CR data.
- Deployment manifests must ensure appropriate `get/list/watch` permissions on `tenants.klutch.anynines.com`.
