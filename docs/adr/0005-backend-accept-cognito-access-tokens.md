# ADR 0005: Backend support for Cognito client_credentials (access token) in non-interactive bind

## Context
- The anynines backend (`/bind-noninteractive`) currently verifies a bearer token using `oidc.Verifier` with a single clientID (from `oidc-config`). It expects an ID token with `aud=<control-plane-client-id>`.
- AWS Cognito’s `client_credentials` grant does **not** return an ID token; it returns an access token (`token_use=access`). That access token typically lacks the `aud` the backend enforces, leading to 401 with errors like `expected audience "<clientID>" got []`.
- We want to support machine-to-machine binding via Cognito without switching providers or flows.

## Decision / Request
- Update the backend to accept Cognito access tokens for `/bind-noninteractive`, while still validating issuer, signature, and scope.
- Key changes:
  - Relax strict ID-token clientID enforcement for this endpoint: either allow multiple clientIDs (ADR 0003) or skip clientID check and rely on issuer + scope.
  - For Cognito, parse and validate access tokens (`token_use=access`), ensure `scope` contains `klutch/bind`, and ensure issuer matches the configured pool.
  - Accept tokens whose `client_id`/`clientid` claim (in Cognito access token) is in an allowlist (from control-plane client and/or Tenant CRDs per ADR 0004).

## Proposed implementation (minimal)
1) In `OIDCServiceProvider`, allow a mode for `/bind-noninteractive` that:
   - Uses `oidc.Config{SkipClientIDCheck:true}` to verify signature/issuer, or manually parses the JWT and validates signature with the provider’s keys.
2) In `handleBindNonInteractive`:
   - After verification, inspect claims:
     - If `token_use` is present, require `token_use == "access"`.
     - Require `scope` (space-separated) contains `klutch/bind`.
     - If an allowlist of client IDs is configured (from control-plane client or Tenant CRD/status), require `client_id`/`clientid` claim to be in that list.
3) Keep current behavior as default when no allowlist is provided to preserve compatibility; enable the relaxed path via config/flag.

## Alternatives
- Switch to an IdP/flow that issues ID tokens for client_credentials (not available in Cognito).
- Force interactive code flow: unsuitable for non-interactive binding.

## Consequences
- Enables non-interactive Cognito bindings without audience failures.
- Slightly relaxes clientID coupling, but still enforces issuer, scope, and optional clientID allowlist for safety.
- Backward compatible if gated by config and default remains current single-client verification when no allowlist is set.
