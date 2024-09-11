---
id: a9s-cli
title: a9s CLI
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

# a9s CLI

anynines provides a command line tool called `a9s` to facilitate application development, devops tasks and interact with selected anynines products.

## Prerequisites

* MacOS / Linux.
* Using the backup/restore feature of a8s PostgreSQL requires an S3 compatible endpoint.
* Install Go (if you want `go env` to identify your OS and arch).
* Install Git.
* Install Docker.
* Install Kubectl.
* Install Kind and/or Minikube.

## Installing the CLI

In order to install the `a9s` CLI execute the following shell script:

```bash
RELEASE=$(curl -L -s https://a9s-cli-v2-fox4ce5.s3.eu-central-1.amazonaws.com/stable.txt); OS=$(go env GOOS); ARCH=$(go env GOARCH); curl -fsSL -o a9s https://a9s-cli-v2-fox4ce5.s3.eu-central-1.amazonaws.com/releases/$RELEASE/a9s-$OS-$ARCH
    
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

### `a8s` Stack
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

### [Go to the a8s Stack documentation](/docs/a9s-cli-a8s/)

### `klutch` Stack
* Install a local Klutch central management cluster using `kind`
* Install Crossplane and the a8s stack on the central management cluster
* Bind resources from a consumer cluster to the management cluster

### [Go to the klutch Stack documentation](/docs/a9s-cli-klutch/)
