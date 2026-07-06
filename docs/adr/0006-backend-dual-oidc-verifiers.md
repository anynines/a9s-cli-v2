# ADR 0006: Dual OIDC verifiers to support Cognito access tokens for non-interactive bind

## Context
- Backend `/bind-noninteractive` currently uses a single OIDC verifier with a fixed `ClientID` (control-plane app client). It expects an ID token with `aud=<clientID>`.
- Cognito `client_credentials` returns an access token (no ID token) with `token_use=access` and typically no matching `aud`, causing 401s even when issuer/client are correct.
- The interactive/browser flow (authorization code) already works with the strict verifier; we don’t want to weaken that path.

## Decision / Request
- Add a second, “lax” verifier for the non-interactive bind endpoint, while keeping the strict verifier for interactive flows.
- The lax verifier should skip `ClientID` enforcement (`SkipClientIDCheck`) but still validate issuer/signature. The handler must then enforce:
  - `token_use == "access"` (Cognito access token)
  - `scope` includes `klutch/bind`
  - `client_id`/`clientid` claim is in an allowlist (control-plane client ID and, when available, per-tenant client IDs from Tenant CRD status per ADR 0004)
- Gate this behavior via config/flag to preserve current behavior by default.

## Suggested implementation (backend-anynines)
- In `http/oidc.go`, add a second verifier: `laxBindVerifier := provider.Verifier(&oidc.Config{SkipClientIDCheck:true})`.
- In `handleBindNonInteractive`, use `laxBindVerifier.Verify(...)`, parse claims, enforce `token_use`, `scope` contains `klutch/bind`, and `client_id` in allowlist.
- Keep the existing strict verifier for `/bind` (interactive) and other paths.

## Benefits
- Unblocks Cognito non-interactive binding (client_credentials) without impacting the working interactive flow.
- Maintains security via issuer/signature/scope/clientID-allowlist checks.
- Backward compatible if gated by config; default can remain strict.
