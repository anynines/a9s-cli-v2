# ADR 0004: Tenant CRD with IdP-specific Controllers (Cognito/Keycloak)

## Context
- Tenants are currently implicit and managed out-of-band (AWS Secrets Manager, CLI flags). This causes drift, audience mismatches (single client ID trusted by backend), and no single source of truth.
- We need to support multiple identity providers (AWS Cognito, Keycloak, etc.) while keeping a consistent tenant model and exposing per-tenant OIDC credentials for workload binding.
- Backend authorization should be able to trust per-tenant client IDs without manual configuration.

## Decision
- Introduce a control-plane-owned `Tenant` CRD as the first-class representation of a Klutch tenant.
- Add IdP-specific controllers (e.g., Cognito Tenant Controller, Keycloak Tenant Controller) that reconcile Tenant resources into the chosen IdP and publish per-tenant OIDC connection details back into Kubernetes.
- The anynines-backend will consume Tenant status (or a derived allowlist) to accept per-tenant client IDs for `/bind-noninteractive`, eliminating the single-client limitation.

## Tenant CRD (proposed)
- `apiVersion: klutch.anynines.com/v1alpha1`
- `kind: Tenant`
- `metadata.name`: human-readable; `spec.tenantUUID`: stable UUID (generated if empty).
- `spec`:
  - `displayName`, `owner` (email/group/ARN), optional `idp` selector (`provider: cognito|keycloak`, hints like `userPoolID`/`realm`).
  - `services`: requested API groups/resources (for default bind request).
- `status` (set by IdP controller):
  - `phase`: Pending | Provisioning | Ready | Error; `conditions`: `IdPReady`, `CredentialsReady`.
  - `oidc`: `issuer`, `tokenURL`, `clientID`, `scope` (e.g., `klutch/bind`), `bindURL`, `bindRequest` (JSON), `secretRef` (for clientSecret).
  - `message`/`reason`.

## Controllers
- **CognitoTenantController**: watches Tenants with `idp.provider == cognito`, ensures shared user pool, resource server/scope, app client per tenant, hosted domain, and writes client ID/secret to a Secret; updates `status.oidc`.
- **KeycloakTenantController**: same contract for Keycloak realms/clients.
- Controllers tag resources per control plane as today (Klutch=ControlPlane, cluster/name/id, etc.).

## Backend integration
- Backend watches Tenants (phase Ready) and builds an allowlist of trusted clientIDs/issuer/scope for `/bind-noninteractive` verification (ties into ADR 0003). Default to single-client behavior if no Tenant data.

## CLI integration
- `a9s create klutch tenant` creates/updates a Tenant CR; waits for Ready and prints `status.oidc` (no need to hit Secrets Manager). `a9s bind klutch workload --tenant-name` reads Tenant+Secret instead of AWS Secrets Manager.

## Alternatives considered
- Keep relying on per-provider secrets only: rejected (no single source of truth, hard to support multiple IdPs, audience drift).
- Encode tenant data solely in Secrets Manager: rejected (no in-cluster visibility for backend/RBAC).

## Consequences / Migration
- Control-plane install deploys Tenant CRD and at least one IdP controller.
- Backend must learn to trust multiple client IDs from Tenants.
- CLI keeps legacy path for a transition period; default moves to Tenant CRD once controllers are present.
