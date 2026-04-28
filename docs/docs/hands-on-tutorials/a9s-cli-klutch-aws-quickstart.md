---
id: hands-on-tutorial-klutch-aws-release-qa
title: "Quick-Start Guide a9s CLI: Klutch on AWS"
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
## Overview

### What you will accomplish

In this tutorial you will learn how to **create control plane and workload clusters** in EKS, fully
equipped **with a PostgreSQL** operator and bound via **klutch-bind**, ready for you to
**provision** a PostgreSQL database instance in the **workload** cluster and **run** it inside the
**control plane** cluster.

### What you will learn

* Install the [a9s CLI](https://github.com/anynines/a9s-cli-v2)
* Create a Klutch Control Plane cluster using EKS
* Create a Tenant to use for binding
* Create a Klutch Workload cluster using EKS
* Create a PostgreSQL database instance
* Create a Service Binding
* Create a backup
* Restore a backup
* Delete the Service Binding
* Delete the restore
* Delete the backup
* Delete the PostgreSQL database instance
* Delete the Klutch Workload cluster
* Delete the Klutch Control Plane cluster

### Prerequisites

* MacOS / Linux
    * Other platforms, including Windows, may work but are currently untested.
* AWS credentials with sufficient permissions in `eu-central-1`.
* A public Route53 hosted zone
* Installed CLIs
  * [a9s CLI](https://github.com/anynines/a9s-cli-v2) $\ge$ v0.16.0 (currently only available as a
    [pre-release version](https://github.com/anynines/a9s-cli-v2/releases/tag/v0.16.0-rc.4))
  * [aws](https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html#getting-started-install-instructions)
    $\ge$ v2.24.20
  * [kubectl](https://kubernetes.io/docs/tasks/tools/#kubectl) $\ge$ v1.27.0
  * [helm](https://helm.sh/docs/intro/install/)
  * [eksctl](https://docs.aws.amazon.com/eks/latest/eksctl/installation.html)
  * [jq](https://jqlang.org/download/)
* Optional but strongly recommended:
  * Fresh AWS account or dedicated test project to avoid collisions.
  * A log file for all commands:

  ```bash
  LOG="manual-qa-$(date +%Y%m%d-%H%M%S).log"
  exec > >(tee -a "$LOG") 2>&1
  ```

## Implementation

In this tutorial you will be using the `a9s` CLI to walk through the complete lifecycle of a local
Klutch Control Plane and Klutch Workload cluster managed via EKS and a PostgreSQL database instance
managed via the a8s PostgreSQL operator.

The `a9s` CLI will guide you through the process while providing you with transparency and ability
to set your own pace. Transparency means that you will see the exact commands to be executed. By
default, non-read commands are executed only after you have confirmed the execution by pressing the
`ENTER` key. This allows you to have a closer look at the command and/or the YAML specifications to
understand what the current step in the tutorial is about.

If all you care about is the result, the `--yes` option will answer all yes-no questions with `yes`,
if you still want the see the commands but don't want to manually approve them all you can use the
`--show-commands` option together with the `--yes` option.
<!-- TODO: change this link to reference the main branch as soon as the changes are merged -->
See [[1]](https://github.com/anynines/a9s-cli-v2/tree/feature-klutch-aws-install) for documentation
and source code of the `a9s` CLI.

## Step 0 - Prerequisites

Make sure that all the necessary CLIs are installed at the correct version and `aws` and `eksctl`
are logged in

```bash
a9s version
aws --version
kubectl version --client
helm version --short
eksctl version
jq --version
aws sts get-caller-identity
eksctl get clusters
```

## Step 1 - Preparation

This tutorial assumes certain environment variables to be set in order for its commands to work:

```bash
export REGION="eu-central-1"
export HOSTED_ZONE="<hosted-zone-name>"
export CP_CLUSTER="my-control-plane-cluster"
export TENANT="my-awesome-tenant"
export WORKLOAD_CLUSTER="my-workload-cluster"
export NS="tutorial"
export PG="my-klutch-pg-instance"
export SB="${PG}-sb"
export BU="${PG}-bu"
export RS="${PG}-rs"
```

The value of `<hosted-zone-name>` can be the name of either an existing Hosted Zone or of a
non-existing subdomain inside an existing Hosted Zone. If a Hosted Zone with the specified name does
not exist, then the `a9s` CLI will create it and attempt to add the necessary records to its parent
zone to make the newly created zone reachable.

## Step 2 - Create control plane cluster

In this section you will create a Control Plane cluster on EKS with a8s PostgreSQL and all its
dependencies as well as the Klutch Tenant operator and the Klutch-Bind backend:

```bash
a9s create cluster klutch control-plane --provider aws \
  --hosted-zone-name "${HOSTED_ZONE}" --cluster-name "${CP_CLUSTER}" \
  --tenant-operator-bind-url "https://klutch-bind.${HOSTED_ZONE}/bind-noninteractive"
```

Currently, the only supported value for the `--provider` option is `aws`, but in the future we plan
on making other cloud infrastructure providers available as well.

> [!Warning] Parent Hosted Zone in a different AWS Account It is possible to use the `a9s` CLI to
> create a new child Hosted Zone, even if the parent Hosted Zone lives in a different AWS account
> than the one the CLI has access to, although this requires human intervention.
>
> If the CLI notices, that no parent Hosted Zone can be detected in the current AWS account for the
> newly created child zone, then it will output a set of DNS name servers looking similar to these:
> `ns-xxxx.awsdns-xx.net. ns-xxxx.awsdns-xx.co.uk. ns-xxxx.awsdns-xx.com. ns-xxxx.awsdns-xx.org`
>
> In order to make the newly created child zone resolvable via DNS you need to open the parent zone
> in the AWS Console, add a record with the name of the child zone and the type `NS` to the parent
> zone's records and paste the DNS name servers from the CLI output into that new record's `value`
> field.

## Step 3 - Create Tenant and verify OIDC secret (Control Plane)

The next step is to create a Klutch Tenant using the following commands:

```bash
aws eks update-kubeconfig --name "${CP_CLUSTER}" --region "${REGION}"
a9s create klutch tenant --tenant-name "${TENANT}"
```

This Tenant is a object managed by a dedicated Operator, the Tenant operator, which runs in the
Control Plane Cluster and maintains a Cognito User Pool as well as a Secrets Manager credential
secret for each Tenant. These two resources make it possible for the Workload cluster to
authenticate to the Control Plane cluster during the binding process.

## Step 4 - Create workload cluster and auto-bind

Now we are ready to create a Klutch Workload cluster using the following command:

```bash
a9s create cluster klutch workload -p aws \
  --tenant-name "${TENANT}" \
  --eks-nodes 1 \
  --cluster-name "${WORKLOAD_CLUSTER} \
  --control-plane-cluster ${CP_CLUSTER}

aws eks update-kubeconfig --name "${WORKLOAD_CLUSTER}" --region "${REGION}"
kubectl create namespace "${NS}"
```

The CLI will now set up a Workload cluster on EKS with the name `$WORKLOAD_CLUSTER` and use the
information in `$TENANT`'s credential secret to bind the Workload Cluster to the Control Plane
cluster named `$CP_CLUSTER`.

## Step 5 - Create Klutch PostgreSQL instance (Workload)

Once both the Control Plane cluster and the Workload cluster are set up, you can create a PostgreSQL
instance managed by the a8s PostgreSQL operator by using the following command:

```bash
a9s create klutch pg instance --name "${PG}" -n "${NS}"
```

> [!Note] The conditions of the `postgresqlinstance` object created by the `a9s` CLI (e.g.
> `READY=true`) do not reflect the actual state of the PostgreSQL instance in the Control Plane
> cluster. Please use the following command to check the actual state of the instance:
>
> ```bash
> kubectl get postgresqlinstances.anynines.com "${PG}" -n "${NS}" -o jsonpath='{.status.managed}'
> ```

## Step 6: Interacting with PostgreSQL

Once you've created a PostgreSQL Service Instance, you can use the `a9s CLI` to interact with it.

### Applying a Local SQL File

Although not the preferred way to load seed data into a production database, during development it
might be handy to execute a SQL file to a PostgreSQL instance. This allows executing one or multiple
SQL statements conveniently.

Download an exemplary SQL file:

```bash
curl https://a9s-cli-v2-fox4ce5.s3.eu-central-1.amazonaws.com/demo_data.sql -o demo_data.sql
```

Executing an SQL file is as simple as using the `--file` option:

```bash
a9s pg apply --file demo_data.sql -i clustered-instance -n tutorial
```

The `a9s CLI` will determine the replication leader, upload, execute and delete the SQL file.

The `--no-delete` option can be used during debugging of erroneous SQL statements as the SQL file
remains in the PostgreSQL Leader's Pod.

```bash
a9s pg apply --file demo_data.sql -i clustered-instance -n tutorial --no-delete
```

With the SQL file still available in the Pod, statements can be quickly altered and re-tested.

### Applying an SQL String

It is also possible to execute a SQL string containing one or several SQL statements by using the
`--sql` option:

```bash
a9s pg apply -i clustered-instance -n tutorial --sql "SELECT COUNT(*) FROM posts"
```

The output of the command will be printed on the screen, for example:

```
Output from the Pod:

count
-------
    10
(1 row)
```

Again, the `pg apply` commands are not meant to interact with production databases but may become
handy during debugging and local development.

Be aware that these commands are executed by the privileged `postgres` user. Schemas (tables)
created by the `postgres` user may not be accessible by roles (users) created in conjunction with
Service Bindings. You will then have to grant access privileges to the Service Binding role.

## Step 7 - Create Klutch service binding (Workload)

After the PostgreSQL instance is created, you can provision a Service Binding with non-privileged
credentials for accessing the instance using the following command:

```bash
a9s create klutch pg servicebinding --name "${SB}" -i "${PG}" -n "${NS}"
```

> [!Note] The conditions of the `servicebinding` object created by the `a9s` CLI (e.g. `READY=true`)
> do not reflect the actual state of the Service Binding in the Control Plane cluster. Please use
> the following command to check the actual state of the instance:
>
> ```bash
> kubectl get servicebindings.anynines.com "${SB}" -n "${NS}" -o jsonpath='{.status.managed}'
> ```

You can then extract the credentials from the Service Binding using the following commands:

```bash
kubectl get secret "${SB}-service-binding" -n "${NS}" -o jsonpath='{.data.database}' | base64 -d; echo
kubectl get secret "${SB}-service-binding" -n "${NS}" -o jsonpath='{.data.instance_service}' | base64 -d; echo
kubectl get secret "${SB}-service-binding" -n "${NS}" -o jsonpath='{.data.username}' | base64 -d; echo
kubectl get secret "${SB}-service-binding" -n "${NS}" -o jsonpath='{.data.password}' | base64 -d; echo
```

> [!Note] Currently the `a9s` CLI deploys no solution for routing traffic from an application
> deployed inside the Workload cluster to the PostgreSQL instance inside the Control Plane cluster.
>
> In the future we do, however, plan on extending the setup done by the `a9s create cluster klutch
> control-plane` command to include a preconfigured deployment of [Envoy
> Gateway](https://gateway.envoyproxy.io/). This will make it possible to easily expose the
> PostgreSQL instances in the Control Plane cluster to outside requests using AWS LoadBalancers.

## Step 8 - Create Klutch backup

After interacting with the PostgreSQL instance you can create a backup of its data using the
following command:

```bash
a9s create klutch pg backup --name "${BU}" -i "${PG}" -n "${NS}"
```

> [!Note] The conditions of the `backup` object created by the `a9s` CLI (e.g. `READY=true`) do not
> reflect the actual state of the Backup in the Control Plane cluster. Please use the following
> command to check the actual state of the backup process:
>
> ```bash
> kubectl get backups.anynines.com "${BU}" -n "${NS}" -o jsonpath='{.status.managed}'
> ```

## Step 9 - Create Klutch restore (new feature)

Restoring a backup works with this command:

```bash
a9s create klutch pg restore --name "${RS}" -b "${BU}" -i "${PG}" -n "${NS}"
```

> [!Note] The conditions of the `restore` object created by the `a9s` CLI (e.g. `READY=true`) do not
> reflect the actual state of the Restore in the Control Plane cluster. Please use the following
> command to check the actual state of the restoration process:
>
> ```bash
> kubectl get restores.anynines.com "${RS}" -n "${NS}" -o jsonpath='{.status.managed}'
> ```

## Step 10 - Delete service binding

Once you are done with a ServiceBinding you can delete it using this command:

```bash
a9s delete klutch pg servicebinding --name "${SB}" -n "${NS}" --wait
```

You can verify its deletion by checking for the absence of servicebinding- and secret-objects in the
tutorial namespace:

```bash
kubectl get servicebindings.anynines.com "${SB}" -n "${NS}" --ignore-not-found
kubectl get secret "${SB}-service-binding" -n "${NS}" --ignore-not-found
```

## Step 11 - Delete restore, backup, and instance

Once you are done with a Restore, Backup or PostgreSQL instance you can delete it using one of these
commands:

```bash
a9s delete klutch pg restore --name "${RS}" -n "${NS}" --wait
a9s delete klutch pg backup --name "${BU}" -n "${NS}" --wait
a9s delete klutch pg instance --name "${PG}" -n "${NS}" --wait
```

You can verify the deletion by checking for the absence of backup-, restore- or
postgresqlinstance-objects in the tutorial namespace:

```bash
kubectl get restores.anynines.com "${RS}" -n "${NS}" --ignore-not-found
kubectl get backups.anynines.com "${BU}" -n "${NS}" --ignore-not-found
kubectl get postgresqlinstances.anynines.com "${PG}" -n "${NS}" --ignore-not-found
```

> [!Note] Currently there is a bug which prevents the a8s Backup Manager from deleting the data of a
> backup from the Minio instance deployed on the Control Plane cluster. A fix for this is currently
> being worked on.

## Step 12 - Delete workload cluster

If you don't need the Workload cluster any more you can delete it using one of these commands:

* minimal cleanup logic - will leave disabled KMS keys behind:

  ```bash
  a9s delete cluster klutch workload -p aws --cluster-name "${WORKLOAD_CLUSTER}"
  ```

* key cleanup logic - will schedule KMS keys used by that cluster for deletion after 7 days, but
  will leave the Hosted Zone used for exposing the Klutch-Bind backend and the ACM certificate used
  for TLS traffic to the backend behind:

  ```bash
  a9s delete cluster klutch workload -p aws --cluster-name "${WORKLOAD_CLUSTER}" --schedule-kms-deletion
  ```

## Step 13 - Delete control plane cluster

If you don't need the Control Plane cluster any more you can delete it using one of these commands:

* minimal cleanup logic - will leave disabled KMS keys, the Hosted Zone used for exposing the
  Klutch-Bind backend and the ACM certificate used for TLS traffic to the backend behind:

  ```bash
  a9s delete cluster klutch control-plane -p aws
  ```

* key cleanup logic - will schedule KMS keys used by that cluster for deletion after 7 days, but
  will leave the Hosted Zone used for exposing the Klutch-Bind backend and the ACM certificate used
  for TLS traffic to the backend behind:

  ```bash
  a9s delete cluster klutch control-plane -p aws --schedule-kms-deletion
  ```

* DNS cleanup logic - will delete the Hosted Zone used for exposing the Klutch-Bind backend and the
  ACM certificate used for TLS traffic to the backend but will leave any KMS keys used by that
  cluster behind in a disabled state:

  ```bash
  a9s delete cluster klutch control-plane -p aws --cleanup-dns-acm --hosted-zone-name "${HOSTED_ZONE}"
  ```

* full cleanup logic - will schedule KMS keys used by that cluster for deletion after 7 days, delete
  the Hosted Zone used for exposing the Klutch-Bind backend and delete the ACM certificate used for
  TLS traffic to the backend:

  ```bash
  a9s delete cluster klutch control-plane -p aws -- schedule-kms-deletion --cleanup-dns-acm --hosted-zone-name "${HOSTED_ZONE}"
  ```
