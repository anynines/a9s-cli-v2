---
id: troubleshooting
title: Troubleshooting
sidebar_position: 4
tags:
  - troubleshooting
  - klutch
  - aws
keywords:
  - troubleshooting
  - cognito
  - hosted zone
  - kube context
  - known issues
---

Known hazards you can hit while following the tutorials, and what each one looks like when it happens.

## Control plane creation appears to hang for 30+ minutes

**Symptom.** `a9s create cluster klutch control-plane` sits without visible progress during the Cognito step. The step is documented as taking about 20 minutes; it runs well past that.

**Cause.** The IPv6 endpoint for AWS Cognito is sometimes very slow to respond, taking minutes per call, while the IPv4 endpoint answers immediately. The Cognito setup can add up to **35 minutes** to control plane creation as a result.

**What to do.** Temporarily disable IPv6 resolution before running the step. On macOS:

```bash
networksetup -setv6off Wi-Fi
```

Re-enable it afterwards:

```bash
networksetup -setv6automatic Wi-Fi
```

The command is not stuck and does not need interrupting. It is waiting on Cognito.

## Hosted zone never becomes resolvable

**Symptom.** Control plane creation prints a set of NS records and then polls, without completing.

**Cause.** The `a9s` CLI created a child hosted zone but could not find its parent in the current AWS account, so it cannot add the delegation records itself. The zone stays unresolvable until they exist.

**What to do.** Add an `NS` record to the parent zone, in the account that holds it, with the child zone's name as the record name and the printed name servers as the value. The CLI polls for up to 30 minutes and continues on its own once the delegation resolves.

Deleting such a zone later is also manual. See the warning on the teardown step of the AWS tutorial.

## Commands act on the wrong cluster

**Symptom.** `kubectl` reports that a resource does not exist when you just created it, or a Klutch resource kind is unrecognised.

**Cause.** The AWS tutorial moves between the Control Plane and Workload clusters several times. Every command's correctness depends on which one your kubeconfig currently points at, and nothing in the shell prompt tells you.

**What to do.** Confirm the current context before running commands against a cluster:

```bash
kubectl config current-context
```

To list the Klutch clusters the CLI knows about and switch between them:

```bash
a9s get clusters klutch
a9s use klutch --cluster-name <name>
```

Remember that Klutch resource kinds only exist in a **Workload** cluster after it has been bound, and that `a9s pg apply` runs against the **Control Plane**.

## Backup data remains in Minio after deletion

**Symptom.** Backups are deleted through the API, but their data is still present in the Minio object store on the Control Plane cluster.

**Cause.** A known issue: the a8s Backup Manager does not remove backup data from Minio when the Backup object is deleted. A fix is in progress.

**What to do.** Nothing is required for the tutorials to work. Be aware that tearing down Klutch resources does not reclaim that storage, and that it survives until the cluster itself is deleted.

## An object says `READY` but nothing works

**Symptom.** A PostgreSQL instance, Service Binding, Backup, or Restore in the Workload cluster reports `READY`, but connecting or using it fails.

**Cause.** `READY` on these objects reflects whether the konnector propagated them to the Control Plane, not whether the underlying resource is usable.

**What to do.** Read `.status.managed` instead. See [`READY` versus `.status.managed`](./concepts/index.md) in Concepts for the full explanation and the exact command.
