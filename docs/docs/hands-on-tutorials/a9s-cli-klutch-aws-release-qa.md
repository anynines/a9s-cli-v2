---
id: hands-on-tutorial-klutch-aws-release-qa
title: Manual QA Tutorial - Klutch AWS Release Gate
tags:
  - a9s CLI
  - klutch
  - aws
  - qa
  - release

keywords:
  - a9s cli
  - klutch
  - aws
  - release gate
  - manual qa
---

# Manual QA Tutorial - Klutch AWS Release Gate

This tutorial is the final manual quality gate before releasing `a9s` CLI changes for the AWS Klutch flow.

It validates the full roundtrip:

1. Create control plane cluster
2. Create tenant
3. Create workload cluster and bind
4. Create PostgreSQL instance
5. Create service binding
6. Create backup (new)
7. Create restore (new)
8. Delete binding
9. Delete restore
10. Delete backup
11. Delete service instance
12. Delete workload cluster
13. Delete control plane cluster

It also validates command-path separation:

- Local a8s operator path: `a9s create pg ...`
- Remote Klutch-managed path: `a9s create klutch pg ...`

## Scope and quality objective

The goal is to prove that the release supports a complete create/use/delete lifecycle without silent leftovers and with clear command semantics.

A release should only proceed if:

- All mandatory steps in this tutorial pass.
- No P0/P1 issues remain open.
- Any accepted leftovers are explicitly documented and approved.

## Prerequisites

- AWS credentials with sufficient permissions in `eu-central-1`.
- A delegated Route53 hosted zone (example: `hub.test.a9s.io`).
- Installed CLIs: `a9s`, `aws`, `kubectl`, `helm`, `eksctl`, `jq`.
- You understand this is a very expensive test (EKS + NAT + ALB + Cognito + Secrets Manager).

Optional but strongly recommended:

- Fresh AWS account or dedicated test project to avoid collisions.
- A log file for all commands:

```bash
LOG="manual-qa-$(date +%Y%m%d-%H%M%S).log"
exec > >(tee -a "$LOG") 2>&1
```

## Naming variables

Use explicit variables to avoid command mistakes:

```bash
export REGION="eu-central-1"
export HOSTED_ZONE="hub.test.a9s.io"
export CP_CLUSTER="klutch-control-plane"
export TENANT="tenant-$(date -u +%y%m%d%H%M%S)-$((RANDOM%900+100))"
export WORKLOAD_CLUSTER=""
export NS="a8s-workload"
export PG="klutch-pg-$(date -u +%y%m%d%H%M%S)"
export SB="${PG}-sb"
export BU="${PG}-bu"
export RS="${PG}-rs"
```

## Step 0 - Preflight checks

```bash
a9s version
aws sts get-caller-identity
kubectl version --client
helm version --short
```

Fail immediately if any command is missing or AWS identity is wrong.

## Step 1 - Create control plane cluster

```bash
a9s create cluster klutch control-plane -p aws --verbose --yes \
  --hosted-zone-name "${HOSTED_ZONE}" \
  --tenant-operator-bind-url "https://klutch-bind.${HOSTED_ZONE}/bind-noninteractive"
```

Verify:

```bash
aws eks describe-cluster --name "${CP_CLUSTER}" --region "${REGION}" --query 'cluster.status' --output text
curl -fsSL "https://klutch-bind.${HOSTED_ZONE}/healthz"
```

Expected:

- EKS cluster status is `ACTIVE`.
- Health endpoint returns JSON with `status: ok`.

## Step 2 - Create tenant and verify OIDC secret

```bash
aws eks update-kubeconfig --name "${CP_CLUSTER}" --region "${REGION}"
a9s create klutch tenant --tenant-name "${TENANT}" --yes
kubectl get tenant "${TENANT}" -n a9s-tenants-operator-system -o yaml
```

Verify secret and token exchange:

```bash
SECRET="klutch/${TENANT}/oidc-client"
aws secretsmanager get-secret-value --secret-id "${SECRET}" --region "${REGION}" --query SecretString --output text | jq .
```

Extract `token_url`, `client_id`, `client_secret`, `scope` and request an access token:

```bash
TOKEN_URL="$(aws secretsmanager get-secret-value --secret-id "${SECRET}" --region "${REGION}" --query SecretString --output text | jq -r .token_url)"
CLIENT_ID="$(aws secretsmanager get-secret-value --secret-id "${SECRET}" --region "${REGION}" --query SecretString --output text | jq -r .client_id)"
CLIENT_SECRET="$(aws secretsmanager get-secret-value --secret-id "${SECRET}" --region "${REGION}" --query SecretString --output text | jq -r .client_secret)"
SCOPE="$(aws secretsmanager get-secret-value --secret-id "${SECRET}" --region "${REGION}" --query SecretString --output text | jq -r .scope)"
curl -fsS -u "${CLIENT_ID}:${CLIENT_SECRET}" \
  --data-urlencode 'grant_type=client_credentials' \
  --data-urlencode "scope=${SCOPE}" \
  "${TOKEN_URL}" | jq -r .access_token | head -c 24
echo
```

Expected:

- Tenant reaches `status.phase: Ready`.
- Secret exists and contains valid OIDC values.
- Token request succeeds.

## Step 3 - Create workload cluster and auto-bind

```bash
CREATE_OUTPUT="$(a9s create cluster klutch workload -p aws --tenant-name "${TENANT}" --eks-nodes 1 --yes)"
echo "${CREATE_OUTPUT}"
WORKLOAD_CLUSTER="$(echo "${CREATE_OUTPUT}" | sed -n 's/.*Generated workload cluster name: \(klutch-workload-cluster-[a-z0-9]\+\).*/\1/p' | head -n1)"
echo "WORKLOAD_CLUSTER=${WORKLOAD_CLUSTER}"
```

Verify:

```bash
aws eks describe-cluster --name "${WORKLOAD_CLUSTER}" --region "${REGION}" --query 'cluster.status' --output text
aws eks update-kubeconfig --name "${WORKLOAD_CLUSTER}" --region "${REGION}"
kubectl get crd | rg "postgresqlinstances.anynines.com|servicebindings.anynines.com|backups.anynines.com|restores.anynines.com"
kubectl create namespace "${NS}" --dry-run=client -o yaml | kubectl apply -f -
```

Expected:

- Workload cluster exists and is active.
- Binding-related CRDs are present.
- Namespace created or already exists.

## Step 4 - Command-path separation checks (critical)

The release introduces/depends on explicit separation:

- Remote Klutch path: `a9s create klutch pg ...`
- Local a8s path: `a9s create pg ...`

Mandatory checks:

1. On the Klutch workload cluster, use only `a9s create klutch pg ...` for service lifecycle.
2. Confirm that your test checklist and team docs do not call local `a9s create pg ...` for the Klutch flow.
3. In a separate local-a8s environment, still verify local commands work unchanged.

## Step 5 - Create Klutch PostgreSQL instance

```bash
a9s create klutch pg instance --name "${PG}" -n "${NS}" --verbose --yes
kubectl get postgresqlinstances.anynines.com "${PG}" -n "${NS}" -o yaml
```

Expected:

- CLI returns success.
- Claim exists and reaches `Ready`.

## Step 6 - Create Klutch service binding

```bash
a9s create klutch pg servicebinding --name "${SB}" -i "${PG}" -n "${NS}" --verbose --yes
kubectl get servicebindings.anynines.com "${SB}" -n "${NS}" -o yaml
kubectl get secret "${SB}-service-binding" -n "${NS}" -o jsonpath='{.data.database}' | base64 -d; echo
kubectl get secret "${SB}-service-binding" -n "${NS}" -o jsonpath='{.data.instance_service}' | base64 -d; echo
kubectl get secret "${SB}-service-binding" -n "${NS}" -o jsonpath='{.data.username}' | base64 -d; echo
kubectl get secret "${SB}-service-binding" -n "${NS}" -o jsonpath='{.data.password}' | base64 -d; echo
```

Expected:

- Service binding claim reaches ready/implemented state.
- Secret exists and all required keys are non-empty.

## Step 7 - Create Klutch backup (new feature)

```bash
a9s create klutch pg backup --name "${BU}" -i "${PG}" -n "${NS}" --verbose --yes
kubectl get backups.anynines.com "${BU}" -n "${NS}" -o yaml
```

Expected:

- Backup command succeeds.
- Backup claim reaches `Ready`.

## Step 8 - Create Klutch restore (new feature)

```bash
a9s create klutch pg restore --name "${RS}" -b "${BU}" -i "${PG}" -n "${NS}" --verbose --yes
kubectl get restores.anynines.com "${RS}" -n "${NS}" -o yaml
```

Expected:

- Restore command succeeds.
- Restore claim reaches `Ready`.

## Step 9 - Delete service binding

```bash
a9s delete klutch pg servicebinding --name "${SB}" -n "${NS}" --verbose --yes --wait
kubectl get servicebindings.anynines.com "${SB}" -n "${NS}" --ignore-not-found
kubectl get secret "${SB}-service-binding" -n "${NS}" --ignore-not-found
```

Expected:

- Binding and generated secret are deleted.

## Step 10 - Delete restore, backup, and instance

```bash
a9s delete klutch pg restore --name "${RS}" -n "${NS}" --verbose --yes --wait
a9s delete klutch pg backup --name "${BU}" -n "${NS}" --verbose --yes --wait
a9s delete klutch pg instance --name "${PG}" -n "${NS}" --verbose --yes --wait
```

Verify absence:

```bash
kubectl get restores.anynines.com "${RS}" -n "${NS}" --ignore-not-found
kubectl get backups.anynines.com "${BU}" -n "${NS}" --ignore-not-found
kubectl get postgresqlinstances.anynines.com "${PG}" -n "${NS}" --ignore-not-found
```

Expected:

- All claims are removed cleanly.

## Step 11 - Delete workload cluster

```bash
a9s delete cluster klutch workload -p aws --cluster-name "${WORKLOAD_CLUSTER}" --yes --really --verbose
```

Verify:

```bash
aws eks describe-cluster --name "${WORKLOAD_CLUSTER}" --region "${REGION}" >/dev/null 2>&1 || echo "workload cluster deleted"
aws ec2 describe-vpcs --region "${REGION}" --filters "Name=tag:Klutch,Values=Workload" "Name=tag:Name,Values=${WORKLOAD_CLUSTER}-vpc" --query 'Vpcs[*].VpcId' --output text
```

Expected:

- Workload cluster deleted.
- Workload VPC no longer exists.

## Step 12 - Delete control plane cluster

```bash
a9s delete cluster klutch control-plane -p aws --verbose --yes --really
```

Verify:

```bash
aws eks describe-cluster --name "${CP_CLUSTER}" --region "${REGION}" >/dev/null 2>&1 || echo "control-plane cluster deleted"
aws ec2 describe-vpcs --region "${REGION}" --filters "Name=tag:Klutch,Values=ControlPlane" "Name=tag:Name,Values=${CP_CLUSTER}-vpc" --query 'Vpcs[*].VpcId' --output text
```

Expected:

- Control-plane cluster deleted.
- Control-plane VPC no longer exists.

## Final integrity checks

Run after cleanup:

```bash
aws eks list-clusters --region "${REGION}" --query 'clusters[?starts_with(@, `klutch-`)]' --output text
aws ec2 describe-vpcs --region "${REGION}" --filters Name=tag:Klutch,Values=ControlPlane,Workload --query 'Vpcs[*].VpcId' --output text
aws route53 list-resource-record-sets --hosted-zone-id <HOSTED_ZONE_ID> --query "ResourceRecordSets[?Name=='klutch-bind.${HOSTED_ZONE}.']"
aws secretsmanager describe-secret --secret-id "klutch/${TENANT}/oidc-client" --region "${REGION}" --query Name --output text
```

Decide and document if remaining resources are intended or not.

## What to look out for during testing

- Wrong command family usage:
  - `a9s create pg ...` (local path) used in Klutch flow by mistake.
  - `a9s create klutch pg ...` should be used for remote flow.
- OIDC secret quality:
  - malformed `token_url` (for example missing domain prefix),
  - invalid client credentials,
  - missing `bind_url` or `bind_request`.
- Slow deletes and stuck dependencies:
  - NAT gateways can take a long time to delete.
  - ELB ENIs can delay VPC teardown.
- Residual billable resources after cleanup:
  - Route53 records,
  - Cognito pools,
  - Secrets Manager secrets,
  - IAM policies/roles.

## Release verdict template

Record one verdict at the end:

- `PASS`: all mandatory steps passed, no P0/P1 issues.
- `PASS WITH KNOWN ISSUES`: no P0/P1 issues, leftovers/risks explicitly accepted.
- `FAIL`: at least one mandatory step failed or unresolved P0/P1 issue.

## Open issues currently known before release

At the time of writing, the following items should be explicitly reviewed before release approval:

1. Control-plane delete can leave a stale `klutch-bind.<hosted-zone>` Route53 CNAME when DNS cleanup flags are not set.
2. Tenant Cognito user pools and Secrets Manager tenant secrets may remain after control-plane deletion.
3. AWS e2e currently focuses on instance and service binding lifecycle; backup/restore lifecycle coverage should be included as first-class release tests for the new Klutch backup/restore commands.

