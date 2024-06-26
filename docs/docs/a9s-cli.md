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
---

# a9s CLI

anynines provides a command line tool called `a9s` to facilitate application development, devops tasks and interact with selected anynines products.

# Prerequisites

* Using the backup/restore feature of a8s PostgreSQL requires an S3 compatible endpoint.
* Install Go (if you want `go env` to identify your OS and arch).
* Install Git.
* Install Docker.
* Install Kubectl.
* Install Kind and/or Minikube.
* Install the [cert-manager CLI](https://cert-manager.io/docs/reference/cmctl/).

# Installing the CLI

In order to install the `a9s` CLI execute the following shell script:

    RELEASE=$(curl -L -s https://a9s-cli-v2-fox4ce5.s3.eu-central-1.amazonaws.com/stable.txt); OS=$(go env GOOS); ARCH=$(go env GOARCH); curl -fsSL -o a9s https://a9s-cli-v2-fox4ce5.s3.eu-central-1.amazonaws.com/releases/$RELEASE/a9s-$OS-$ARCH
        
    sudo chmod 755 a9s
    sudo mv a9s /usr/local/bin

This will download the `a9s` binary suitable for your architecture and move it to `/usr/local/bin`.
Depending on your system you have to adjust the `PATH` variable or move the binary to a folder that's already in the `PATH`.

# Using the CLI

    a9s

# Creating a Local a8s Postgres Cluster

Create a local Kubernetes cluster using `Minikube` or `Kind`, install a8s PostgreSQL including its dependencies ready for **local development of applications requiring PostgreSQL** and/or **experimentation with a8s Postgres** by issuing the command:

    a9s create cluster a8s

Recommended is 12 GB of free memory for the creation of three cluster nodes with each 4 GB. The number of nodes and memory size can be adjusted.

## Cold-Run

When creating a cluster for the first time, a few setup steps will have to be taken which need to be performed only once:

1. Setting up a working directory for the use with the `a9s` CLI.
2. Configuring the access credentials for the S3 compatible object store which is needed if you intend to use the backup/restore feature of a8s Postgres.
3. Cloning deployment resources required by the `a9s` CLI to create a cluster.

### Setting Up a Working Directory

The working directory is where are `a9s` CLI related resources will go. This includes `yaml` specifications being cloned from remote repositories, but also those generated by the `a9s` CLI for your convenience.

Once established, the working directory is stored in the `~/.a8s` configuration file.

Establishing a working directory is simple. Create an empty folder at a place of your choice such as `~/a9s-workspace`. The simplest way to use this folder, is to change into the folder and execute the `a9s create cluster a8s` from it. The CLI will automatically propose the current directory as the working directory. 

    cd ~
    mkdir a9s-workspace
    cd a9s-workspace

    a9s create cluster a8s

Alternatively, provide a custom working directory at the corresponding prompt.

### (Optional) Configuring the Backup Store

Defaults:

* The default **infrastructure region** is `eu-central-1`.
* The default **backup provider** is `AWS`.
* The default **bucket** is `a8s-backups`.

See `a9s create cluster a8s --help` for the defaults of your particular CLI version and list of configuration options.

When prompted provide the following pieces of information:

* `ACCESS KEY ID`
* `SECRET KEY`

**Skipping the Backup Store Configuration**:

In case you don't want to use the backup/restore function, paste arbitrary strings as `ACCESS KEY ID` and `SECRET KEY`. Backup and restore jobs will fail but the managing service instances and service bindings will work.


## Skip Checking Prerequisites

It is possible to skip the verification of prerequisites. This includes skipping the search for: required shell commands, a running Docker daemon and a running Kubernetes cluster.

In order to skip precheck use the `--no-precheck` option:

    a9s create cluster a8s --no-precheck

## Number of Kubernetes Nodes

Specifying the number of Nodes in the Kubernetes cluster:

    a9s create cluster a8s --cluster-nr-of-nodes 1

## Cluster Memory

Specifying the memory of **each** Node of the Kubernetes cluster:


    a9s create cluster a8s --cluster-memory 4gb

## Deployment Version

The deployment version refers to the version of manifests used for installing software. Deployment versions are managed by anynines in a Git repository. The deployment version option allows you to select a particular version of the deployment manifests identified by **Git tags**.


Select a particular release by providing the `--deployment-version` parameter:

    a9s create cluster a8s --deployment-version v0.3.0

Use:

    a9s create cluster a8s --deployment-version latest

To get the latest, untagged version of the deployment manifests.

## Kubernetes Provider

When creating a Kubernetes cluster, the mechanism to manage the cluster can be selected by specifying the `--provider` option:

    a9s create cluster a8s -p kind 
    a9s create cluster a8s -p minikube (default)

Follow the instructions to learn about available sub commands.

## Backup Infrastructure Region

When using the backup and restore functionality, a backup infrastructure region must be specified by using the `--backup-region` option:

    a9s create cluster a8s --backup-region us-east-1

**Note**: By default, an existing `backup-config.yaml` will be used. Hence, if you intend to change
your backup config, remove the existing `backup-config.yaml`, first:

    rm a8s-deployment/deploy/a8s/backup-config/backup-store-config.yaml

## Unattended Mode

It is possible to skip all yes-no questions by **enabling the unattended mode** by passing the `-y` or `--yes` flag:

    a9s create cluster a8s --yes

## Printing the Working Directory

The working directory is stored in the `~/.a8s` configuration file. The working directory contains all resources downloaded and generated by the `a9s` CLI.

To print the working directory execute:


    a9s cluster pwd

# a8s PostgreSQL

A selected subset of the a8s PostgreSQL features are available through the `a9s` CLI.

## Creating a PostgreSQL Service Instance

Creating a service instance with the name `sample-pg-cluster`:

    a9s create pg instance --name sample-pg-cluster

The generated YAML specification will be stored in the `usermanifests`.

### Creating PostgreSQL Service Instance YAML Without Applying it

    a9s create pg instance --name sample-pg-cluster --no-apply

The generated YAML specification will be stored in the `usermanifests` but `kubectl apply` won't be executed.

### Creating a Custom PostgreSQL Service Instance

The command:

    a9s create pg instance --api-version v1beta3 --name my-pg --namespace default --replicas 3 --req
uests-cpu 200m --limits-memory 200Mi --service-version 14 --volume-size 2Gi

Will generate a YAML spec called `usermanifests/my-pg-instance.yaml` with the following content:

```yaml
apiVersion: postgresql.anynines.com/v1beta3
kind: Postgresql
metadata:
  name: my-pg
spec:
  replicas: 3
  resources:
    limits:
      memory: 200m
    requests:
      cpu: 200m
  version: 14
  volumeSize: 2Gi
```

## Deleting a PostgreSQL Service Instance

Deleting a service instance with the name `sample-pg-cluster`:

    a9s delete pg instance --name sample-pg-cluster

Or by providing an explicit namespace:

    a9s delete pg instance --name sample-pg-cluster -n default

**Note**: If the service instance doesn't exist, a warning is printed and the command exists with the
return code `0` as the desired state of the service instance being delete is reached.


## Applying a SQL File to a PostgreSQL Service Instance

Uploading a SQL file, executing it using `psql` and deleting the file can be done with:

    a9s pg apply --file /path/to/sql/file --instance-name sample-pg-cluster

The file is uploaded to the current primary pod of the service instance. 

**Note**: Ensure that, during the execution of the command, there is no change of the primary node for a given clustered service instance as otherwise the file upload may fail or target the wrong pod.

Use `--yes` to skip the confirmation prompt.

    a9s pg apply --file /path/to/sql/file --instance-name sample-pg-cluster --yes

Use `--no-delete` to leave the file in the pod:

    a9s pg apply --file /path/to/sql/file --instance-name sample-pg-cluster --no-delete

## Applying a SQL Statement to a PostgreSQL Service Instance

Applying a SQL statement on the primary pod of a PostgreSQL service instance can be accomplished with:

    a9s pg apply -i solo --sql "select count(*) from posts" --yes

## Creating a Backup of a PostgreSQL Service Instance

    a9s create pg backup --name sample-pg-cluster-backup-1 -i sample-pg-cluster-1

## Restoring a Backup of PostgreSQL Service Instance

    a9s create pg restore --name sample-pg-cluster-restore-1 -b sample-pg-cluster-backup-1 -i sample-pg-cluster-1

## Creating a PostgreSQL Service Binding

A Service Binding is an entity facilitating the secure consumption of a service instance.
By creating a service instance, a Postgres user is created along with a corresponding Kubernetes Secret.

    a9s create pg servicebinding --name sb-clustered-1 -i clustered

Will therefore create a Kubernetes Secret named `sb-clustered-1-service-binding` and provide the following 
keys containing everything an application needs to connect to the PostgreSQL service instance:

- `database`
- `instance_service`
- `password`
- `username`


# Cleaning Up

In order to delete the cluster run:

    a9s delete cluster a8s

**Note**: This will not delete config files.

Config files are stored in the cluster working directory.

They can be removed with:

    rm -rf $( a9s cluster pwd )