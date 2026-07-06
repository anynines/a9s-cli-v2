# Feature Request: anynines-backend Support for Per-Tenant Scopes/Clients

## Context
- Klutch is moving toward per-tenant app clients and per-tenant scopes (e.g., `klutch/bind:<tenant_uuid>`) on a single Cognito pool.
- The backend currently assumes a single client/scope and fails with 401 when tokens use per-tenant scopes/clients.

## Request
- Update the anynines-backend OIDC auth to:
  - Accept tokens from the shared issuer with varying client_ids (per tenant).
  - Accept scopes matching `klutch/bind:*` or otherwise not reject per-tenant scopes.
  - Continue to enforce issuer, scope, and tenant isolation (e.g., via `tenant_id` claim).

## Benefit
- Enables true multi-tenant isolation via per-tenant scopes/clients while keeping a single issuer.
- Avoids having to fall back to a single shared scope/client just to satisfy the backend.
