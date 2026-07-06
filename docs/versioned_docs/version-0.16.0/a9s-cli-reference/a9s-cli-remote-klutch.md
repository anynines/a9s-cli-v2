---
id: a9s-cli-klutch-on-aws
title: Remote Klutch Stack (AWS)
tags:
  - a9s cli
  - klutch
  - aws
  - eks
keywords:
  - a9s cli
  - klutch
  - aws
  - eks
  - control plane
  - workload
  - tenant
  - postgresql
---

Create and manage a Klutch Control Plane and Workload clusters on AWS EKS. The Remote Klutch Stack
provisions production-grade EKS clusters with Crossplane, the a8s PostgreSQL operator, the
Klutch-Bind backend, and tenant-aware OIDC (Cognito). Workload clusters are bound to the Control
Plane via non-interactive OIDC flows, enabling developers to provision data services declaratively
from their own clusters.

## Prerequisites

- [General prerequisites](./index.md#prerequisites) are met.
- An AWS account with sufficient permissions (EKS, VPC, IAM, Cognito, Secrets Manager, Route53, ACM,
  KMS).
- Current AWS CLI profile (either `default` or set via `AWS_PROFILE` environment variable)
  configured (`aws configure`) with credentials for the target account.
- `kubectl` installed and available in `PATH`.
- `helm` installed and available in `PATH`.
- A public Route53 hosted zone (required for TLS and the bind backend endpoint).

## Global Flags

The following flags are available for all subcommands:

| Flag | Description | Default |
| ------ | ------------- | --------- |
| `-y`, `--yes` | Skip any confirmation prompts. | `false` |
| `-v`, `--verbose` | Prints additional information such as the StdOut and StdErr of shell commands. | `false` |
| `--show-commands` | Prints the shell commands used under the hood by the CLI as they are executed. | `false` |

## Commands

### 1. create cluster klutch control-plane

#### Usage

```bash
a9s create cluster klutch control-plane -p aws --hosted-zone-name <zone> [options]
```

#### Options

| Flag | Description | Default |
| ------ | ------------- | --------- |
| `-p`, `--provider` | Provider (only `aws` supported) | `aws` |
| `-c`, `--cluster-name` | EKS cluster name | `klutch-control-plane` |
| `--hosted-zone-name` | Route53 hosted zone (absolute domain name). **Required** unless `--no-apply`. | - |
| `--region` | The region in which to create the EKS cluster and all the networking infrastructure | eu-central-1 |
| `--no-apply` | Provision the EKS cluster without installing control plane components. | `false` |
| `--dry-run` | Show planned AWS resources without creating them. | `false` |
| `--eks-node-type` | EC2 instance type for worker nodes. | `t3a.xlarge` |
| `--eks-nodes` | Number of worker nodes (min/max/desired). | `3` |
| `--host` | Ingress host override (defaults to API server host). | - |
| `--ingress-port` | Port for the ALB ingress. | `443` |
| `--acm-certificate-arn` | ACM certificate ARN for HTTPS. Auto-provisioned if omitted. | - |
| `--oidc-provider` | OIDC provider (`cognito` or `dex`). | `cognito` (when `-p aws`) |
| `--oidc-issuer-url` | OIDC issuer URL (required if `--oidc-provider cognito` is manually set). | - |
| `--oidc-client-id` | OIDC client ID (required if `--oidc-provider cognito` is manually set). | - |
| `--oidc-client-secret` | OIDC client secret (required if `--oidc-provider cognito` is manually set). | - |
| `--oidc-callback-url` | OIDC callback URL. | `https://<host>/callback` |
| `--tenant-operator-image` | Tenant operator container image override. | - |
| `--tenant-operator-chart` | Tenant operator Helm chart (OCI URL). | - |
| `--tenant-operator-chart-version` | Chart version (for OCI charts). | - |
| `--tenant-operator-role-arn` | IAM role ARN for the tenant operator (IRSA). | - |
| `--tenant-operator-region` | Region for tenant operator AWS calls. | control-plane region |
| `--tenant-operator-bind-url` | Bind URL for tenant operator config. | derived |
| `--tenant-operator-bind-request` | Bind request JSON for tenant operator config. | all exported services |
| `--klutch-bind-backend-img` | Override backend image as `<repo>:<tag>`. | - |
| `--klutch-bind-backend-img-url` | Override backend image repository. **Note**: only checked if `--klutch-bind-backend-img` flag is not set. | - |
| `--klutch-bind-backend-img-tag` | Override backend image tag. **Note**: only checked if `--klutch-bind-backend-img` flag is not set. | - |

#### Description

Creates an EKS cluster and installs the full Klutch control plane stack:

- VPC with public/private subnets, NAT gateways, and Internet Gateway
- EKS cluster with managed nodegroup and gp3 default StorageClass
- AWS ALB Ingress Controller
- Crossplane and anynines configuration packages
- The a8s PostgreSQL operator (backup, restore, service binding)
- The Klutch-Bind backend (fronted by ALB with ACM TLS)
- The Tenant Operator (IRSA-enabled, managing Cognito app clients)
- Route53 DNS records and ACM certificate (auto-provisioned if not supplied)

The cluster uses a KMS key for secret encryption. By default, the CLI creates a new key for each
cluster. If the `KEY_ID` environment variable is populated the CLI will use the key with the
specified ID instead.

#### Example

```bash
a9s create cluster klutch control-plane -p aws \
  --hosted-zone-name klutch.example.com \
  --eks-nodes 3 \
  --eks-node-type t3a.xlarge
```

---

### 2. create cluster klutch workload

#### Usage

```bash
a9s create cluster klutch workload -p aws [options]
```

#### Options

| Flag | Description | Default |
| ------ | ------------- | --------- |
| `-p`, `--provider` | Provider (only `aws` supported). | `aws` |
| `-c`, `--cluster-name` | Workload cluster name. Auto-generated if omitted and `WORKLOAD_CLUSTER_NAME` environment variable is not set. Naming pattern for auto-generation: `klutch-workload-cluster-` followed by 4 randomly chosen bytes encoded in hexadecimal. | `WORKLOAD_CLUSTER_NAME` environment variable if set, otherwise random |
| `--tenant-name` | Tenant name for auto-binding. | - |
| `--tenant-secret-name` | Explicit Secrets Manager secret name. | `klutch/<tenant>/oidc-client` |
| `--region` | The region in which to create the EKS cluster and all the networking infrastructure | eu-central-1 |
| `--tenant-region` | AWS region for the tenant secret. | CONTROL_PLANE_CLUSTER_REGION or eu-central-1 |
| `--bind-request-file` | Override the tenant's stored bind request JSON. | - |
| `--control-plane-cluster` | Control plane cluster name for CA lookup. | `klutch-control-plane` |
| `--eks-node-type` | EC2 instance type for worker nodes. | `t3a.xlarge` |
| `--eks-nodes` | Number of worker nodes. | `3` |
| `--dry-run` | Show planned resources without creating them. | `false` |

#### Description

Creates an EKS workload cluster. When `--tenant-name` is provided, the CLI automatically retrieves
OIDC credentials from Secrets Manager and binds the workload cluster to the control plane after
provisioning.

The cluster uses a KMS key for secret encryption. By default, the CLI creates a new key for each
cluster. If the `KEY_ID` environment variable is populated the CLI will use the key with the
specified ID instead.

#### Example

```bash
a9s create cluster klutch workload -p aws \
  --tenant-name team-alpha \
  --cluster-name klutch-workload-team-alpha
```

---

### 3. create klutch tenant

#### Usage

```bash
a9s create klutch tenant --tenant-name <name> [options]
```

#### Options

| Flag | Description | Default |
| ------ | ------------- | --------- |
| `--tenant-name` | Name for the tenant. **Required**. | - |
| `--region` | AWS region for Cognito. | CONTROL_PLANE_CLUSTER_REGION or eu-central-1 |
| `--store-secret` | Store credentials in Secrets Manager. | `true` |
| `--secret-name` | Secrets Manager secret name. | `klutch/<tenant>/oidc-client` |
| `--force` | Overwrite existing tenant secret. | `false` |
| `--bind-request-file` | Path to bind request JSON. | all exported services |

#### Description

Applies a Tenant Custom Resource to the control-plane cluster. The Tenant Operator reconciles this
into a Cognito app client, resource server scope, and Secrets Manager secret containing OIDC
credentials, bind URL, and bind request. Workload clusters use these credentials for non-interactive
binding.

#### Example

```bash
a9s create klutch tenant --tenant-name team-alpha
```

---

### 4. apply klutch control-plane

#### Usage

```bash
a9s apply klutch control-plane --hosted-zone-name <zone> [options]
```

#### Options

| Flag | Description | Default |
| ------ | ------------- | --------- |
| `--hosted-zone-name` | Route53 hosted zone (FQDN). **Required**. | - |
| `--host` | Ingress host override. | API server host |
| `--ingress-port` | Ingress port. | `443` |
| `--acm-certificate-arn` | ACM certificate ARN. Auto-provisioned if omitted. | - |
| `-p`, `--provider` | Provider (`aws` enables IRSA/ALB addons). | - |
| `-c`, `--cluster-name` | Cluster name for addon deployment. | `klutch-control-plane` |
| `--oidc-provider` | OIDC provider (`cognito` or `dex`). | `cognito` (when `-p aws`) |
| `--oidc-issuer-url` | OIDC issuer URL. | - |
| `--oidc-client-id` | OIDC client ID. | - |
| `--oidc-client-secret` | OIDC client secret. | - |
| `--oidc-callback-url` | OIDC callback URL. | `https://<host>/callback` |
| `--tenant-operator-image` | Tenant operator image override. | - |
| `--tenant-operator-chart` | Tenant operator Helm chart (OCI URL). | - |
| `--tenant-operator-chart-version` | Chart version (for OCI charts). | - |
| `--tenant-operator-role-arn` | IAM role ARN for IRSA. | - |
| `--tenant-operator-region` | Region for tenant operator. | control-plane region |
| `--tenant-operator-bind-url` | Bind URL override. | derived |
| `--tenant-operator-bind-request` | Bind request JSON override. | all exported services |
| `--klutch-bind-backend-img` | Override backend image as `<repo>:<tag>`. | - |
| `--klutch-bind-backend-img-url` | Override backend image repository. **Note**: only checked if `--klutch-bind-backend-img` flag is not set. | - |
| `--klutch-bind-backend-img-tag` | Override backend image tag. **Note**: only checked if `--klutch-bind-backend-img` flag is not set. | - |

#### Description

Installs the Klutch control plane components onto an **existing** Kubernetes cluster (typically an
EKS cluster already provisioned via `create cluster klutch control-plane --no-apply` or by other
means). Installs Crossplane, the a8s stack, the Klutch-Bind backend, ALB ingress, and the Tenant
Operator.

#### Example

```bash
a9s apply klutch control-plane \
  --hosted-zone-name klutch.example.com \
  -p aws \
  --cluster-name my-existing-cluster
```

---

### 5. bind klutch workload

#### Usage

```bash
a9s bind klutch workload [options]
```

#### Options

##### Regular Flags

These flags can be used in interactive and non-interactive mode:

| Flag | Description | Default |
| ------ | ------------- | --------- |
| `--tenant-name` | Tenant name to load OIDC credentials from Secrets Manager. | - |
| `--tenant-secret-name` | Explicit secret name. | `klutch/<tenant>/oidc-client` |
| `--tenant-region` | AWS region for the tenant secret. | CONTROL_PLANE_CLUSTER_REGION or eu-central-1 |
| `--control-plane` | Control plane bind endpoint URL (overrides tenant secret). | - |
| `--control-plane-cluster` | Control plane cluster name for CA lookup. | `klutch-control-plane` |
| `--bind-request-file` | Path to bind request JSON (overrides tenant secret). | - |
| `--oidc-client-id` | OIDC client ID override. | from tenant secret |
| `--oidc-client-secret` | OIDC client secret override. | from tenant secret |
| `--oidc-token-url` | OIDC token URL override. | from tenant secret |
| `--oidc-scope` | OIDC scope override. | from tenant secret |
| `--kubeconfig` | Path to workload cluster kubeconfig. | current context |
| `--context` | Workload cluster kubeconfig context. | current context |
| `--konnector-image` | Override the konnector image. | - |
| `--skip-konnector` | Skip konnector deployment. | `false` |
| `--write-kubeconfig` | Write control-plane kubeconfig to path. | - |
| `--dry-run` | Dry-run mode. | `false` |
| `--interactive-bind` | Use browser-based interactive klutch-bind flow. | `false` |

##### Interactive-Mode-Only Flags

These flags can only be used together with the `--interactive-bind` flag:

| Flag | Description | Default |
| ------ | ------------- | --------- |
| `--bind-arg` | can be used to pass a list of additional arguments to the klutch-bind plugin | - |
| `-o`,`--output` | output mode for the klutch-bind plugin | - |

#### Description

Binds a workload cluster to a Klutch control plane. By default uses the **non-interactive** flow:
retrieves OIDC credentials from the tenant secret, exchanges them for an access token, calls
`/bind-noninteractive` on the backend, deploys the konnector, and applies the bind request.

With `--interactive-bind`, opens a browser for the interactive klutch-bind flow (requires
`--control-plane`).

#### Example

```bash
# Non-interactive (recommended)
a9s bind klutch workload --tenant-name team-alpha

# Interactive
a9s bind klutch workload --interactive-bind --control-plane https://klutch-bind.example.com/exports
```

---

### 6. delete cluster klutch control-plane

#### Usage

```bash
a9s delete cluster klutch control-plane -p aws [options]
```

#### Options

| Flag | Description | Default |
| ------ | ------------- | --------- |
| `-p`, `--provider` | Provider (only `aws`). | `aws` |
| `-c`, `--cluster-name` | Cluster name. | `klutch-control-plane` |
| `--dry-run` | Show planned deletions without executing. | `false` |
| `--region` | The region of the EKS cluster and networking infrastructure to delete | eu-central-1 |
| `--cleanup-dns-acm` | Delete Route53 records/zone and ACM certificate. | `false` |
| `--delete-dns-zone` | Delete Route53 hosted zone and records. | `false` |
| `--delete-acm-certificate` | Delete ACM certificate. | `false` |
| `--hosted-zone-name` | Hosted zone name (required with DNS flags). | - |
| `--acm-certificate-arn` | ACM certificate ARN to delete. | auto-discovered |
| `--cleanup-orphans` | Remove Klutch-tagged orphaned AWS resources (EIPs). | `false` |
| `--schedule-kms-deletion` | Schedule KMS key deletion (7-day wait). | `false` |
| `--really` | Confirm destructive deletion (only effective if `--yes` is also set). | `false` |

#### Description

Deletes the Klutch control plane EKS cluster and all tagged AWS infrastructure: VPC, subnets, NAT
gateways, Internet Gateway, route tables, security groups, IAM roles, managed nodegroups, ALB
controller, and the EKS cluster itself. Optionally cleans up Route53 DNS, ACM certificates, and
orphaned resources.

#### Example

```bash
a9s delete cluster klutch control-plane -p aws \
  --cleanup-dns-acm --hosted-zone-name klutch.example.com \
  --schedule-kms-deletion --really --yes
```

---

### 7. delete cluster klutch workload

#### Usage

```bash
a9s delete cluster klutch workload -p aws --cluster-name <name> [options]
```

#### Options

| Flag | Description | Default |
| ------ | ------------- | --------- |
| `-p`, `--provider` | Provider (only `aws`). | `aws` |
| `-c`, `--cluster-name` | Workload cluster name. **Required** if `WORKLOAD_CLUSTER_NAME` environment variable is not set. | `WORKLOAD_CLUSTER_NAME` environment variable if set |
| `--dry-run` | Show planned deletions without executing. | `false` |
| `--region` | The region of the EKS cluster and networking infrastructure to delete | eu-central-1 |
| `--cleanup-orphans` | Remove Klutch-tagged orphaned resources. | `false` |
| `--schedule-kms-deletion` | Schedule KMS key deletion (7-day wait). | `false` |
| `--really` | Confirm destructive deletion (only effective if `--yes` is also set). | `false` |

#### Description

Deletes the specified Klutch workload EKS cluster and its tagged AWS infrastructure.

#### Example

```bash
a9s delete cluster klutch workload -p aws \
  --cluster-name klutch-workload-team-alpha --really --yes
```

---

### 8. delete klutch control-plane

#### Usage

```bash
a9s delete klutch control-plane
```

#### Description

Removes the Klutch control plane **components** (Dex/Cognito config, backend, ingress, Crossplane
Helm release) from the current kube context without deleting the underlying cluster. Use this to
uninstall Klutch from an existing EKS cluster.

---

### 9. delete klutch tenant

#### Usage

```bash
a9s delete klutch tenant <tenant-name> [options]
```

#### Options

| Flag | Description | Default |
| ------ | ------------- | --------- |
| `--region` | AWS region for Secrets Manager. | CONTROL_PLANE_CLUSTER_REGION or eu-central-1 |
| `--secret-name` | Explicit secret name. | `klutch/<tenant>/oidc-client` |

#### Description

Deletes the Secrets Manager secret holding the tenant's OIDC credentials.

#### Example

```bash
a9s delete klutch tenant team-alpha
```

---

### 10. get klutch tenant / get klutch tenants

#### Usage

```bash
a9s get klutch tenants [options]
a9s get klutch tenant <tenant-name> [options]
```

**Options** (`tenants`):

| Flag | Description | Default |
| ------ | ------------- | --------- |
| `--region` | AWS region for Secrets Manager. | CONTROL_PLANE_CLUSTER_REGION or eu-central-1 |
| `--prefix` | Secret name prefix to filter tenants. | `klutch/` |

**Options** (`tenant`):

| Flag | Description | Default |
| ------ | ------------- | --------- |
| `--region` | AWS region for Secrets Manager. | CONTROL_PLANE_CLUSTER_REGION or eu-central-1 |
| `--secret-name` | Explicit secret name. | `klutch/<tenant>/oidc-client` |

#### Description

- `get klutch tenants` - lists all Klutch tenant secrets in Secrets Manager.
- `get klutch tenant <name>` - displays the OIDC credentials (issuer, client ID, client secret,
  scope) for a specific tenant.

#### Example

```bash
a9s get klutch tenants
a9s get klutch tenant team-alpha
```

---

### 11. Klutch PostgreSQL Resources

Manage Klutch-managed PostgreSQL claims on a **workload cluster** that is bound to a control plane.

#### `create klutch pg instance`

```bash
a9s create klutch pg instance --name <name> [options]
```

| Flag | Description | Default |
| ------ | ------------- | --------- |
| `--name` | Instance claim name. **Required**. | - |
| `-n`, `--namespace` | Namespace. | `default` |
| `--service` | Service name. | `a9s-postgresql13` |
| `--plan` | Plan name. | `postgresql-single-nano` |
| `--expose` | Exposure mode. | `Internal` |
| `--composition` | Composition name. | `a8s-postgresql` |
| `--no-apply` | Render manifest only. | `false` |
| `--wait` | Wait for Ready condition. | `true` |
| `--wait-timeout` | Timeout for `--wait`. | `30m` |

#### `create klutch pg servicebinding`

```bash
a9s create klutch pg servicebinding --name <name> --service-instance <instance> [options]
```

| Flag | Description | Default |
| ------ | ------------- | --------- |
| `--name` | Binding claim name. **Required**. | - |
| `-i`, `--service-instance` | Instance to bind to. **Required**. | - |
| `-n`, `--namespace` | Namespace. | `default` |
| `--service-instance-type` | Instance type. | `postgresql` |
| `--composition` | Composition name. | `a8s-servicebinding` |
| `--no-apply` | Render manifest only. | `false` |
| `--wait` | Wait for implementation. | `true` |
| `--wait-timeout` | Timeout for `--wait`. | `15m` |

#### `create klutch pg backup`

```bash
a9s create klutch pg backup --name <name> --service-instance <instance> [options]
```

| Flag | Description | Default |
| ------ | ------------- | --------- |
| `--name` | Backup claim name. **Required**. | - |
| `-i`, `--service-instance` | Instance to back up. **Required**. | - |
| `-n`, `--namespace` | Namespace. | `default` |
| `--service-instance-type` | Instance type. | `postgresql` |
| `--composition` | Composition name. | `a8s-backup` |
| `--no-apply` | Render manifest only. | `false` |
| `--wait` | Wait for Ready. | `true` |
| `--wait-timeout` | Timeout for `--wait`. | `30m` |

#### `create klutch pg restore`

```bash
a9s create klutch pg restore --name <name> --backup <backup> --service-instance <instance> [options]
```

| Flag | Description | Default |
| ------ | ------------- | --------- |
| `--name` | Restore claim name. **Required**. | - |
| `-b`, `--backup` | Backup claim to restore from. **Required**. | - |
| `-i`, `--service-instance` | Target instance. **Required**. | - |
| `-n`, `--namespace` | Namespace. | `default` |
| `--service-instance-type` | Instance type. | `postgresql` |
| `--composition` | Composition name. | `a8s-restore` |
| `--no-apply` | Render manifest only. | `false` |
| `--wait` | Wait for Ready. | `true` |
| `--wait-timeout` | Timeout for `--wait`. | `30m` |

#### `delete klutch pg instance|servicebinding|backup|restore`

```bash
a9s delete klutch pg instance --name <name> [-n <namespace>] [--wait] [--wait-timeout <duration>]
a9s delete klutch pg servicebinding --name <name> [-n <namespace>] [--wait]
a9s delete klutch pg backup --name <name> [-n <namespace>] [--wait]
a9s delete klutch pg restore --name <name> [-n <namespace>] [--wait]
```

All delete subcommands accept `--name` (required), `-n`/`--namespace` (default `default`), `--wait`
(default `false`), and `--wait-timeout` (default `15m`).
