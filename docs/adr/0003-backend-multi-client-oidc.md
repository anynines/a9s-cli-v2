# ADR 0003: Backend OIDC support for multiple client IDs (per-tenant app clients)

## Context
- The Klutch control plane (backend-kube-bind “anynines backend”) currently verifies OIDC ID tokens with a single client ID configured via the `oidc-config` secret (`oidc-issuer-client-id`). In `backend-anynines/http/oidc.go` the verifier is created as `provider.Verifier(&oidc.Config{ClientID: clientID})`, and `/bind-noninteractive` enforces that verifier.
- In the Cognito-based multi-tenant flow, each workload tenant gets its own app client ID/secret in the same user pool. Tokens issued for these per-tenant clients have `aud = <tenant-client-id>`.
- The backend logs show 401 Unauthorized due to audience mismatch: `expected "3g14gp2d39kieebqqtplso3vlt" got []` when a tenant client token is presented. This blocks non-interactive binding with per-tenant credentials.
- Keycloak docs assume a single client shared by all tenants, so the audience matches. Cognito multi-client support requires backend changes.

## Decision / Request
- Add backend support for multiple acceptable client IDs (audiences) for OIDC verification on `/bind-noninteractive`.
- Prefer pulling the allowed client IDs from first-class Tenant resources in the control plane (Tenant CRD + IdP-specific controllers such as Cognito/Keycloak):
  - Tenants in phase Ready publish `status.oidc.{issuer,tokenURL,clientID,scope,bindURL,bindRequest}` and a `secretRef` for client secret.
  - Backend watches Tenants and builds an allowlist of clientIDs (and issuers/scopes) it trusts for bind requests.
- If Tenant CRD is not yet available, allow configuration of a static list of client IDs (env/flag/ConfigMap) as an interim step.
- Implementation options (any of these would unblock multi-client):
  1) Configure the OIDC verifier to allow any clientID in the allowlist (from Tenants or static config). If the library cannot take multiple IDs, create verifiers per clientID and try them.
  2) Set `SkipClientIDCheck` and enforce issuer match + `aud ∈ allowlist` + scope contains `klutch/bind` in the handler.
  3) Register per-tenant client IDs via Tenant controller and expose them to the backend as a ConfigMap/CRD for trust.

## Rationale
- Per-tenant app clients in Cognito provide tenant isolation and rotation without sharing a global client secret. The backend currently blocks this because of strict single-audience verification.
- Supporting multiple client IDs (or relaxed clientID check with scope enforcement) restores compatibility with the multi-tenant Cognito flow while keeping issuer validation intact.

## Acceptance Criteria
- `/bind-noninteractive` accepts tokens issued by per-tenant app clients in the same issuer when they include the required scope (`klutch/bind`) and have a clientID in the allowlist (Tenant status or static config).
- Audience mismatch errors no longer occur when using per-tenant Cognito or Keycloak clients.
- Backend defaults remain compatible: if no Tenant data or allowlist is provided, it falls back to the current single-client behavior.
