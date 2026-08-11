---
id: concepts-index
title: Concepts
sidebar_position: 1
tags:
  - klutch
  - concepts
  - architecture
keywords:
  - klutch
  - control plane
  - workload cluster
  - tenant
  - bind
  - konnector
  - crossplane claim
---

The rest of this documentation uses these terms as command arguments and object kinds. Read this page first if any of them are new. The tutorials assume them.

## Control Plane cluster and Workload cluster

Klutch splits responsibilities across two Kubernetes clusters.

The **Control Plane cluster** runs the machinery: Crossplane, the a8s PostgreSQL operator, the Klutch-Bind backend, and the Tenant operator. Data service instances are actually provisioned here.

A **Workload cluster** is where application teams work. It runs no operators of its own. Once bound to a Control Plane, it gains the Klutch resource kinds (`postgresqlinstances.anynines.com` and friends), and a developer submits requests there without needing access to the Control Plane.

A single Control Plane serves many Workload clusters. The `a9s` CLI creates each with a different command: `a9s create cluster klutch control-plane` and `a9s create cluster klutch workload`.

:::note

The Klutch project's own documentation calls a Workload cluster an **app cluster**. Same thing; this documentation uses "Workload cluster" throughout, matching the CLI's command names.

:::

## Tenant

A **Tenant** represents an entity that will bind Workload clusters to a Control Plane: a team, a project, or an environment.

Creating one with `a9s create klutch tenant` makes the Tenant operator:

1. Create a **Cognito app client** with a `client_credentials` grant and a `klutch/bind` scope.
2. Store the resulting OIDC credentials in **AWS Secrets Manager**.

Those credentials are then consumed automatically when a Workload cluster binds, which is why binding on AWS needs no interactive login.

## bind and the konnector

**Binding** is what connects a Workload cluster to a Control Plane. It establishes an OIDC-authenticated session and installs the konnector.

The **konnector** is a lightweight agent that runs in the Workload cluster and maintains that session. It carries submissions from the Workload cluster to the Control Plane and brings status back. It is the reason a Workload cluster can offer data service APIs while running no operators itself.

The sync is always initiated by the Workload cluster. The Control Plane does not reach into it.

## Crossplane Claim

A **Claim** is the namespace-scoped resource a developer submits to ask for a data service instance: a `PostgresqlInstance`, a `ServiceBinding`, a `Backup`, a `Restore`.

Crossplane reconciles each Claim into a cluster-scoped **Composite Resource (XR)** on the Control Plane, and a provider turns that into a real instance. You normally interact only with Claims.

When you run `a9s create klutch pg instance`, you are creating a Claim.

## a8s objects and Klutch remote claims

Both tutorials provision PostgreSQL, but through different mechanisms, and the object kinds differ.

The **local a8s tutorial** talks to the a8s PostgreSQL operator directly. Objects are the operator's own kinds (`postgresqls.postgresql.anynines.com`, `servicebindings.anynines.com`), and they live in the one cluster you created.

The **Klutch AWS tutorial** goes through Crossplane. Objects are Claims (`postgresqlinstances.anynines.com`) submitted in the Workload cluster and reconciled on the Control Plane.

Similar names, different mechanisms. A command from one tutorial will not work in the other's cluster.

## `READY` versus `.status.managed`

This distinction decides whether you can tell success from failure, and it catches people out.

For Klutch resources in a **Workload** cluster (PostgreSQL instances, Service Bindings, Backups, Restores), the `READY` condition reflects **propagation status**: whether the konnector has synced the object to the Control Plane. It does **not** report whether the underlying database, binding, backup, or restore is actually ready.

An object can show `READY` while the thing it asked for is still being created, or has failed.

Here is the trap in practice. Seconds after creating a PostgreSQL instance, `kubectl get` already reports it ready:

```
NAME                    SYNCED   READY   CONNECTION-SECRET   AGE
my-klutch-pg-instance   True     True                        10s
```

But the database does not exist yet:

```bash
kubectl get postgresqlinstances.anynines.com "${PG}" -n "${NS}" \
  -o jsonpath='{.status.managed}'
```

```json
{"clusterStatus":"Pending"}
```

Roughly two minutes later, the same command returns a healthy instance with one running replica:

```json
{"clusterStatus":"Running","readyReplicas":1}
```

`READY` never changed. Only `.status.managed` did.

Wait for `clusterStatus` to reach `Running` before using the instance. The same field applies to `servicebindings.anynines.com`, `backups.anynines.com`, and `restores.anynines.com`, though the keys inside it differ by resource kind.

Anywhere the tutorials tell you to wait for something, `.status.managed` is the field that answers the question.
