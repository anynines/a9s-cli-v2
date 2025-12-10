# ADR 0001: Cognito Multi-Tenant OIDC for Klutch

## Status
Accepted

## Context
- Klutch control-plane needs OIDC for non-interactive bind flow.
- Tenants range from namespaces to fleets of clusters with multiple users/groups.
- The backend expects a single issuer; per-tenant issuers caused 401s (tokens from different pools rejected).
- We need tenant isolation and straightforward automation for tenant creation/binding.

## Decision
- Use a single Cognito user pool per control-plane (one issuer).
- Create one app client per tenant in that pool; each tenant has unique `tenant_uuid` and human-friendly `tenant_name`.
- Define per-tenant scopes on the shared resource server: `klutch/bind:<tenant_uuid>`.
- Tokens include `tenant_uuid` (used as `tenant_id`) and the per-tenant scope; backend enforces issuer + scope/tenant.
- Tenant secrets store issuer, token_url, client_id/secret, `bind_url`, and `bind_request`.

## Alternatives Considered
- Per-tenant Cognito pools (multiple issuers): rejected due to backend verifier complexity and 401s.
- Shared scope only (`klutch/bind`) + tenant claim: simpler but relies entirely on claim enforcement; per-tenant scopes provide stronger isolation signaling.
- Multiple issuers trusted by backend: added complexity and coordination.

## Consequences
- Backend can keep a single issuer and audience model; tenant isolation is conveyed via per-tenant scopes and `tenant_uuid` claims.
- Tenant creation reuses the control-plane pool; no new pools per tenant.
- Operational overhead is limited to adding app clients and scopes within one pool.
- Workload bind uses the tenant’s client credentials and per-tenant scope; tokens are accepted by the backend without issuer mismatch.
