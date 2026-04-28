---
id: hands-on-tutorial-klutch-aws-quickstart
tags:
  - a9s CLI
  - klutch
  - aws
  - tutorial

keywords:
  - a9s cli
  - klutch
  - aws
  - eks
  - postgresql
  - control plane
  - workload cluster
---

# Klutch on AWS — Quick-Start Tutorial

## Overview

### What you will accomplish

By the end of this tutorial you will have:

1. Created a **Klutch Control Plane** cluster on AWS EKS, running the a8s PostgreSQL operator, the
   Klutch-Bind backend, and the Tenant operator.
2. Created a **Klutch Workload** cluster on AWS EKS and bound it to the Control Plane.
3. Provisioned a **PostgreSQL instance** from the Workload cluster, backed up its data, and restored
   the backup.
4. Torn everything down cleanly.

### What you will learn

* How the `a9s` CLI orchestrates EKS clusters, Crossplane, and Klutch components.
* How Tenants, OIDC credentials, and non-interactive binding connect a Workload cluster to a Control
  Plane.
* How to manage the full lifecycle of a PostgreSQL instance (create, query, bind, back up, restore,
  delete) through Klutch remote claims.
* How to clean up all AWS resources created during the tutorial.

### Estimated duration

Allow **60–90 minutes**. The two EKS cluster creations (~20 minutes each) account for most of the
wall-clock time.

## Prerequisites

### Operating system

* macOS or Linux. Other platforms may work but are untested.

### AWS account

* Credentials with sufficient permissions in `eu-central-1`.
* A **public Route53 hosted zone** either to be used as-is for exposing the Klutch-Bind backend or
  to act as parent for a newly created Hosted Zone which exposes the backend (see Step 2).
* A fresh AWS account or a dedicated test project is strongly recommended to avoid resource
  collisions.

### Required CLIs

| Tool | Minimum version | Install guide |
|------|-----------------|---------------|
| [a9s CLI](https://github.com/anynines/a9s-cli-v2) | v0.16.0 (currently only available as pre-release) | [Releases](https://github.com/anynines/a9s-cli-v2/releases) |
| [aws](https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html) | v2.24.20 | AWS docs |
| [kubectl](https://kubernetes.io/docs/tasks/tools/#kubectl) | v1.27.0 | Kubernetes docs |
| [helm](https://helm.sh/docs/intro/install/) | — | Helm docs |
| [eksctl](https://docs.aws.amazon.com/eks/latest/eksctl/installation.html) | — | AWS docs |
| [jq](https://jqlang.org/download/) | — | jqlang.org |

### Optional: log all terminal output

Capturing a full session log helps with debugging and review:

```bash
LOG="klutch-aws-tutorial-$(date +%Y%m%d-%H%M%S).log"
exec > >(tee -a "$LOG") 2>&1
```
## A note on verifying readiness for a8s-managed resources
The `READY` condition on the PostgreSQL instances, Service Bindings, Backups and Restores in the
**Workload** cluster reflects the propagation status, **not** the readiness of the underlying
database/binding/backup/restore in the **Control Plane**. To check the true state use the
`.status.managed` field of the objects in the **Workload** cluster, such as here:

```bash
kubectl get postgresqlinstances.anynines.com "${PG}" -n "${NS}" \
  -o jsonpath='{.status.managed}'
```

## Step 1 — Verify prerequisites

Run the following commands and confirm that every tool reports a version at or above the minimum
listed above:

```bash
a9s version
aws --version
kubectl version --client
helm version --short
eksctl version
jq --version
```

Verify that your AWS credentials are active and that `eksctl` can reach the API:

```bash
aws sts get-caller-identity
eksctl get clusters
```

**Expected result:** `aws sts get-caller-identity` prints your account ID and ARN. `eksctl get
clusters` either lists existing clusters or returns an empty table — both are fine.

## Step 2 — Set environment variables

All commands in this tutorial reference the variables below. Set them once at the start of your
session:

```bash
export REGION="eu-central-1"
export HOSTED_ZONE="<your-hosted-zone-name>"
export CP_CLUSTER="my-control-plane-cluster"
export TENANT="my-tenant"
export WORKLOAD_CLUSTER="my-workload-cluster"
export NS="tutorial"
export PG="my-klutch-pg-instance"
export SB="${PG}-sb"
export BU="${PG}-bu"
export RS="${PG}-rs"
```

| Variable | Purpose |
|----------|---------|
| `REGION` | AWS region for all resources. |
| `HOSTED_ZONE` | DNS name used by the Klutch-Bind backend. Can be an existing hosted zone or a new subdomain of one. |
| `CP_CLUSTER` | Name of the EKS Control Plane cluster. |
| `TENANT` | Name of the Klutch Tenant (maps to a Cognito app client). |
| `WORKLOAD_CLUSTER` | Name of the EKS Workload cluster. |
| `NS` | Kubernetes namespace for PostgreSQL resources in the Workload cluster. |
| `PG` / `SB` / `BU` / `RS` | Names for the PostgreSQL instance, Service Binding, Backup, and Restore. |

> [!Note]Hosted zone creation
> If a Route53 hosted zone matching `HOSTED_ZONE` does not exist, the
> `a9s` CLI creates one automatically and attempts to add the required NS records in the parent
> zone. If the parent zone is in a different AWS account, the CLI prints the NS records for you to
> add manually — see the callout in Step 3.

## Step 3 — Create the control plane cluster

> [!Note] Note
>
> This step queries the AWS Cognito API to create and configure a user pool for the control plane
> cluster.
>
> The IPv6 endpoint for AWS Cognito can be very slow to react sometimes, in certain cases taking
> multiple minutes to respond while the IPv4 endpoint responds immediately. Because of this issue,
> the Cognito setup can add up to **35 minutes** to the Control Plane cluster creation.
>
> It is therefore advised to temporarily switch off IPv6 resolution (e.g. via `networksetup
> -setv6off Wi-Fi` on MacOs, this can be reversed by running `networksetup -setv6automatic Wi-Fi`)
> before executing this step.
>
> After this step it can be safely enabled again.

This command provisions an EKS cluster and installs the a8s PostgreSQL operator, Crossplane, the
Klutch-Bind backend, and the Tenant operator:

```bash
a9s create cluster klutch control-plane --provider aws \
  --hosted-zone-name "${HOSTED_ZONE}" \
  --cluster-name "${CP_CLUSTER}" \
  --tenant-operator-bind-url "https://klutch-bind.${HOSTED_ZONE}/bind-noninteractive"
```

**What to expect:**

* The CLI frequently uses shell commands to reach its goals. By default, it shows each non-read
  shell command before executing it, and you must press `Enter` to confirm each command. Add `--yes`
  to skip confirmations without seeing which specific commands are executed, or `--yes
  --show-commands` to skip confirmations while still seeing the commands.
* EKS cluster creation takes approximately **15 minutes**. The remaining component installations add
  another 5–10 minutes.
* The command exits when all pods in the cluster are ready.

> [!WARNING] Parent hosted zone in a different AWS account
>
> If the CLI cannot find a parent hosted zone in the current AWS account for the newly created child
> zone, it prints a set of NS records similar to:
>
> ```
> ns-xxxx.awsdns-xx.net.
> ns-xxxx.awsdns-xx.co.uk.
> ns-xxxx.awsdns-xx.com.
> ns-xxxx.awsdns-xx.org.
> ```
>
> To make the child zone resolvable you must add an `NS` record to the parent zone (in the other
> account's Route53 console) with the child zone's name as the record name and these name servers as
> the value. The CLI will poll until the delegation is resolvable (up to 30 minutes) before
> continuing.

## Step 4 — Create a tenant

Switch your kubeconfig to the Control Plane cluster, then create a Tenant:

```bash
aws eks update-kubeconfig --name "${CP_CLUSTER}" --region "${REGION}"
a9s create klutch tenant --tenant-name "${TENANT}"
```

A **Tenant** represents an entity (team, project, environment) that will bind Workload clusters to
this Control Plane. Under the hood the Tenant operator:

1. Creates a **Cognito app client** with `client_credentials` grant and a `klutch/bind` scope.
2. Stores the resulting OIDC credentials in **AWS Secrets Manager**.

These credentials are consumed automatically during the workload binding in the next step.

## Step 5 — Create the workload cluster

Create a Workload cluster and bind it to the Control Plane in a single command:

```bash
a9s create cluster klutch workload -p aws \
  --tenant-name "${TENANT}" \
  --eks-nodes 1 \
  --cluster-name "${WORKLOAD_CLUSTER}" \
  --control-plane-cluster "${CP_CLUSTER}"
```

Once the cluster is ready, update your kubeconfig and create the namespace for this tutorial:

```bash
aws eks update-kubeconfig --name "${WORKLOAD_CLUSTER}" --region "${REGION}"
kubectl create namespace "${NS}"
```

**What to expect:** The CLI creates the EKS cluster (~20 minutes), retrieves the Tenant's OIDC
credentials from Secrets Manager, and performs a **non-interactive bind** to the Control Plane.
After binding, Klutch CRDs (e.g. `postgresqlinstances.anynines.com`) become available in the
Workload cluster.

## Step 6 — Provision a PostgreSQL instance

With both clusters running and bound, create a PostgreSQL instance from inside the **Workload
cluster**:

```bash
a9s create klutch pg instance --name "${PG}" -n "${NS}"
```

Wait until the value of the `.status.managed` field of the newly created `postgresqlinstance`
indicates the instance is running before proceeding.

## Step 7 — Interact with PostgreSQL

Once the PostgreSQL instance is running you can interact with it from inside the **Control Plane**
cluster.

```bash
aws eks update-kubeconfig --name "${CP_CLUSTER}" --region "${REGION}"
```

### Apply a SQL file

Download an example SQL file and execute it against the instance:

```bash
curl -fsSL https://a9s-cli-v2-fox4ce5.s3.eu-central-1.amazonaws.com/demo_data.sql \
  -o demo_data.sql

a9s pg apply --file demo_data.sql -i "${PG}" -n "${NS}"
```

The CLI determines the replication leader, uploads the file, executes it, and removes it from the
pod. Use `--no-delete` to keep the file in the pod for iterative debugging.

### Apply an inline SQL statement

```bash
a9s pg apply -i "${PG}" -n "${NS}" --sql "SELECT COUNT(*) FROM posts"
```

**Expected output:**

```
Output from the Pod:

 count
-------
    10
(1 row)
```

> [!NOTE]
>
> The `pg apply` commands execute as the privileged `postgres` user. Schemas created this way may
> not be accessible by roles provisioned through Service Bindings — you will need to `GRANT`
> privileges explicitly.

## Step 8 — Create a service binding

Switch back to the Workload cluster:

```bash
aws eks update-kubeconfig --name "${WORKLOAD_CLUSTER}" --region "${REGION}"
```

A Service Binding provisions non-privileged credentials for application access to the PostgreSQL
instance from inside the **Workload** cluster:

```bash
a9s create klutch pg servicebinding --name "${SB}" -i "${PG}" -n "${NS}"
```

Verify readiness:

```bash
kubectl get servicebindings.anynines.com "${SB}" -n "${NS}" \
  -o jsonpath='{.status.managed}'
```

### Inspect the credentials

```bash
kubectl get secret "${SB}-service-binding" -n "${NS}" \
  -o jsonpath='{.data.database}' | base64 -d; echo

kubectl get secret "${SB}-service-binding" -n "${NS}" \
  -o jsonpath='{.data.instance_service}' | base64 -d; echo

kubectl get secret "${SB}-service-binding" -n "${NS}" \
  -o jsonpath='{.data.username}' | base64 -d; echo

kubectl get secret "${SB}-service-binding" -n "${NS}" \
  -o jsonpath='{.data.password}' | base64 -d; echo
```

> [!NOTE] The `a9s` CLI does not yet deploy a network routing solution between Workload and Control
> Plane clusters. An Envoy Gateway integration is planned for a future release to expose PostgreSQL
> instances via AWS load balancers.

## Step 9 — Back up and restore (Workload cluster)

### Create a backup

```bash
a9s create klutch pg backup --name "${BU}" -i "${PG}" -n "${NS}"
```

Verify:

```bash
kubectl get backups.anynines.com "${BU}" -n "${NS}" \
  -o jsonpath='{.status.managed}'
```

### Restore the backup

```bash
a9s create klutch pg restore --name "${RS}" -b "${BU}" -i "${PG}" -n "${NS}"
```

Verify:

```bash
kubectl get restores.anynines.com "${RS}" -n "${NS}" \
  -o jsonpath='{.status.managed}'
```

## Step 10 — Clean up resources

Delete resources in reverse dependency order. The `--wait` flag blocks until deletion is confirmed.
Start in the **Workload** cluster and switch over to the **Control Plane** cluster after the
Workload cluster is deleted.

### Delete the service binding

```bash
a9s delete klutch pg servicebinding --name "${SB}" -n "${NS}" --wait
```

Verify removal:

```bash
kubectl get servicebindings.anynines.com "${SB}" -n "${NS}" --ignore-not-found
kubectl get secret "${SB}-service-binding" -n "${NS}" --ignore-not-found
```

### Delete restore, backup, and instance

```bash
a9s delete klutch pg restore --name "${RS}" -n "${NS}" --wait
a9s delete klutch pg backup --name "${BU}" -n "${NS}" --wait
a9s delete klutch pg instance --name "${PG}" -n "${NS}" --wait
```

Verify removal:

```bash
kubectl get restores.anynines.com "${RS}" -n "${NS}" --ignore-not-found
kubectl get backups.anynines.com "${BU}" -n "${NS}" --ignore-not-found
kubectl get postgresqlinstances.anynines.com "${PG}" -n "${NS}" --ignore-not-found
```

> [!NOTE]
>
> There is a known issue where the a8s Backup Manager does not delete backup data from the
> Minio instance on the Control Plane cluster. A fix is in progress.

## Step 11 — Delete the workload cluster

```bash
a9s delete cluster klutch workload -p aws \
  --cluster-name "${WORKLOAD_CLUSTER}" \
  --schedule-kms-deletion
```

The `--schedule-kms-deletion` flag schedules KMS keys created for this cluster for deletion after 7
days (AWS does not allow immediate KMS key deletion). Without this flag, disabled KMS keys are left
behind.

## Step 12 — Delete the control plane cluster

Choose a cleanup level depending on which AWS resources you want to remove. The table below
summarises what each flag does:

| Flag | Effect |
|------|--------|
| *(no cleanup flags set)* | Deletes the EKS cluster and VPC resources. Leaves disabled KMS keys, the hosted zone, and the ACM certificate. |
| `--schedule-kms-deletion` | Also schedules KMS keys for deletion after 7 days. |
| `--delete-acm-certificate` | Also deletes the ACM certificate. |
| `--delete-dns-zone --hosted-zone-name "${HOSTED_ZONE}"` | Also deletes the hosted zone. |
| `--cleanup-dns-acm --hosted-zone-name "${HOSTED_ZONE}"` | Also deletes the ACM certificate and the hosted zone. |

**Recommended — full cleanup:**

```bash
a9s delete cluster klutch control-plane -p aws \
  --cluster-name "${CP_CLUSTER}" \
  --schedule-kms-deletion \
  --cleanup-dns-acm \
  --hosted-zone-name "${HOSTED_ZONE}"
```

## Summary

You have completed the full Klutch-on-AWS lifecycle:

* **Created** a Control Plane cluster with the a8s PostgreSQL operator, Crossplane, and Klutch
  components.
* **Created** a Tenant and a bound Workload cluster.
* **Provisioned** a PostgreSQL instance, created a Service Binding, and performed
  backup/restore — all from the Workload cluster via Klutch remote claims.
* **Cleaned up** all resources in the correct dependency order.

### Further reading

* [a9s CLI source and documentation](https://github.com/anynines/a9s-cli-v2)
* [Klutch concepts](/docs/a9s-cli-klutch)
* [Local a8s PostgreSQL tutorial](/docs/hands-on-tutorials/hands-on-tutorial-a8s-pg-a9s-cli)
