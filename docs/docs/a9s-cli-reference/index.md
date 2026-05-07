---
id: a9s-cli-reference-index
title: a9s CLI Reference
tags:
  - a9s cli
  - a9s hub
  - a9s data services
  - a8s data services
  - a9s postgres
  - a8s postgres
  - data service
  - introduction
  - kubernetes
  - minikube
  - kind
  - klutch
keywords:
  - a9s cli
  - a9s hub
  - a9s platform
  - a9s data services
  - a8s data services
  - a9s postgres
  - a8s postgres
  - data service
  - introduction
  - postgresql
  - kubernetes
  - minikube
  - kind
  - klutch
---

anynines provides a command line tool called `a9s` to facilitate application development, devops tasks and interact with selected anynines products.

## Prerequisites

* MacOS / Linux.
* Using the backup/restore feature of a8s PostgreSQL requires an S3 compatible endpoint.
* Install Git.
* Install Docker.
* Install Kubectl.
* Install Kind and/or Minikube.

## Installing the CLI

In order to install the `a9s` CLI execute the following shell commands:

```bash
OS=$(go env GOOS); ARCH=$(go env GOARCH)
# if you don't have Go installed on your system you can use the following fallback
# OS=$(uname -s | tr '[:upper:]' '[:lower:]'); ARCH=$(uname -m | sed 's/x86_64/amd64/' | sed 's/aarch64/arm64/')

curl -sSL https://github.com/anynines/a9s-cli-v2/releases/download/v0.16.0/a9s-cli-v2_${OS}_${ARCH}.tar.gz | tar -xzf - a9s

sudo chmod 755 a9s
sudo mv a9s /usr/local/bin
```

This will download the `a9s` binary suitable for your architecture and move it to `/usr/local/bin`.
Depending on your system you have to adjust the `PATH` variable or move the binary to a folder that's already in the `PATH`.

## Using the CLI

```bash
a9s
```

## Use Cases

The `a9s` CLI can be used to install and use the following stacks:

### `a8s` Stack (Local)

* Install a local Kubernetes cluster (`minikube` or `kind`).
* Install the [cert-manager](https://cert-manager.io/).
* Install a local Minio object store for storing Backups.
* Install the a8s PostgreSQL Operator PostgreSQL supporting
  * creating dedicated PostgreSQL clusters with
    * synchronous and asynchronous streaming replication.
    * automatic failure detection and automatic failover.
  * backup and restore capabilities storing backups in an S3 compatible object store such as AWS S3 or Minio.
  * ability to easily create database users and Kubernetes Secrets by using the Service Bindings abstraction
* Easily apply `.sql` files and SQL commands to PostgreSQL clusters.

### [Go to the a8s Stack documentation](./a9s-cli-local-a8s.md)

### `klutch` Stack (Local)

* Install a local Klutch Control Plane Cluster using `kind`
* Install Crossplane and the a8s stack on the Control Plane Cluster
* Bind resources from an App Cluster to the Control Plane Cluster

### [Go to the local klutch Stack documentation](./a9s-cli-local-klutch.md)

### `klutch` Stack (AWS)

* Create a production-grade EKS-based Klutch Control Plane with Crossplane, the a8s PostgreSQL operator, and tenant-aware OIDC (Cognito).
* Create EKS workload clusters and bind them to the Control Plane via non-interactive OIDC flows.
* Manage tenants with Cognito app clients and Secrets Manager credentials.
* Provision and manage PostgreSQL instances, backups, restores, and service bindings declaratively from workload clusters.

### [Go to the remote klutch Stack (AWS) documentation](./a9s-cli-remote-klutch.md)

## General Klutch Commands

### `get clusters klutch` / `get klutch clusters`

#### Usage

```bash
a9s get clusters klutch
a9s get klutch clusters
```

#### Description

Lists Klutch-related clusters (control-plane and workload) found in the current kubeconfig. Displays
cluster name and active/inactive status.

#### Example

```bash
a9s get clusters klutch

 ℹ️  Detecting Klutch clusters...
 ╭───────────────────────────────────────╮
 │  Cluster                Status        │
 │ ───────────────────────────────────── │
 │                                       │
 │ ───────────────────────────────────── │
 │  klutch-control-plane   ✅ active     │
 │  kind-klutch-app        ✅ active     │
 ╰───────────────────────────────────────╯
```

---

### `use cluster klutch` / `use klutch`

#### Usage

```bash
a9s use cluster klutch --cluster-name <name>
a9s use klutch --cluster-name <name>
```

#### Options

| Flag | Description | Default |
| ------ | ------------- | --------- |
| `-c`, `--cluster-name` | Klutch cluster name to switch to. **Required**. | - |

#### Description

Switches the current kubectl context to the specified Klutch cluster. Performs substring matching
against available contexts.

#### Example

```bash
a9s use klutch --cluster-name klutch-control-plane
```

---

### `estimate-cost cluster klutch`

#### Usage

```bash
a9s estimate-cost cluster klutch -p aws [options]
```

#### Options

| Flag | Description | Default |
| ------ | ------------- | --------- |
| `-p`, `--provider` | Provider (only `aws`). **Required**. | - |
| `--region` | AWS region to price. | `eu-central-1` |
| `--instance-type` | EC2 instance type for worker nodes. | `t3a.xlarge` |
| `--desired-nodes` | Desired worker node count. | `3` |
| `--min-nodes` | Minimum worker node count. | `3` |
| `--max-nodes` | Maximum worker node count. | `5` |
| `--node-disk-gib` | Root volume size per node (gp3, GiB). | `80` |
| `--nat-gateways` | Number of NAT gateways. | `3` |
| `-o`, `--output` | Output format: `table` or `json`. | `table` |
| `--pricing-region` | AWS region for the Pricing API. | `us-east-1` |

#### Description

Pulls real-time AWS list prices and calculates hourly and monthly cost estimates for the Klutch
cluster bill of materials (EKS control plane fee, worker nodes, root volumes, NAT gateways, KMS
key).

#### Example

```bash
a9s estimate-cost cluster klutch -p aws --region us-west-2 --desired-nodes 5
```
