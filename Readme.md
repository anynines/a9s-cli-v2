# a9s CLI V2

# Development

## Gitflow

This repo is using [gitflow](https://nvie.com/posts/a-successful-git-branching-model/).

## Makefile

There's a `Makefile` to help building and running the cli during development.

## Build

    make build

The binary can be found in `bin/a9s`.

# Using the CLI

    a9s

## Creating a Demo Environment

    a9s create demo a8s

### Skip Checking Prerequisites

In order to skip the verification of necessary commands, a running Docker daemon and a configured Kubernetes cluster:

    a9s create demo a8s --no-precheck

### Number of Kubernetes Nodes

    a9s create demo a8s --cluster-nr-of-nodes 1

### Cluster Memory
    a9s create demo a8s --cluster-memory 12gb

### Deployment Version

Select a particular release by providing the `--deployment-version` parameter:

    a9s create demo a8s --deployment-version v0.3.0

Use

    a9s create demo a8s --deployment-version latest

To get the latest, untagged version of the deployment manifests.

### Kubernetes Provider

If you want to select a particular Kubernetes provider:

    a9s create demo a8s -p kind 
    a9s create demo a8s -p minikube (default)

Follow the instructions to learn about available sub commands.

### Backup Infrastructure Region

    a9s create demo a8s --backup-region us-east-1

**Note**: By default, an existing `backup-config.yaml` will be used. Hence, if you intend to change
your backup config, remove the existing `backup-config.yaml`, first:

    rm a8s-deployment/deploy/a8s/backup-config/backup-store-config.yaml

## Unattended Mode

It is possible to skip all yes-no questions by **enabling the unattended mode** by passing the `-y` or `--yes` flag:

    a9s create demo a8s --yes

## Printing the Demo Working Directory

    a9s demo pwd

## Creating a Service Instance

Creating a service instance with the name `sample-pg-cluster`:

    a9s create pg instance --name sample-pg-cluster

The generated YAML specification will be stored in the `usermanifests`.

### Creating Service Instance YAML Without Applying it

    a9s create pg instance --name sample-pg-cluster --no-apply

The generated YAML specification will be stored in the `usermanifests` but `kubectl apply` won't be executed.

### Creaging a Custom Service Instance

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
``````

## Deleting a Service Instance

Deleting a service instance with the name `sample-pg-cluster`:

    a9s delete pg instance --name sample-pg-cluster

## Creating a Backup of a Service Instance

    a9s create pg backup --name sample-pg-cluster-backup-1 -i sample-pg-cluster-1

## Cleaning Up

In order to delete the Demo cluster run:

    a9s delete demo a8s

**Note**: This will not delete config files stored.

Config files are stored in the demo working directory.

They can be removed with:

    rm -rf $( a9s demo pwd )

# Design Principles / Ideals

* The CLI should not need a tight synchronization with product releases.
    * The release of a new a8s Postgres version, for example, should be working with an existing CLI version.

# Backlog


* Create commnand `a9s create pg backup --name $INSTANCE_NAME`
    * DONE: BackupToYAML implemented
    * DONE: Add command, generate yaml file and optionally execute it

* Extend `a9s create pg instance` to generate a YAML manifest given the params `--name`, `--replicas`, `--volume-size`, `--version`, `--requests-cpu` and `--limits-memory`
    * Establish parameters
    * Create struct
    * Generate yaml - Filename = `"a8s-pg-instance-" + instance_name + ".yaml"
    * Add `-o yaml` flag to generate yaml output, print yaml to screen but do not execute the command, 

* Sub command to delete all demo resources.
    * Remove everything (incl. config files)
        * e.g. `a9s demo delete --all`

* When executing a9s create demo a8s for the first time, the infrastructure-region should be queried as a user input instead of being a default-parameter. The probability is too high that the user choses a non-viable default option instead of providing a valid region.

* Question: Should the demo a8s-pg execute the entire demo or just install the operator? Other commands could be: 
    * a8s-pg 
        * `create`
            * It's more idiomatic in Kubernetes for the verb to be the first command: `kubectl get pods` vs `kubectl pod get`.
        * `a9s pg instance`
            * `create`
                * `a9s pg instance create --isolation pod` > a8s PG
                * `a9s pg create instance --isolation pod`
                * `a9s pg instance create --isolation vm` > a9s PG
        * `a9s pg service-binding`
            * `a9s pg binding` 
            * `a9s pg sb`
        * `a9s pg backup`
        * `a9s pg restore`
    * a8s-pg-instance 
    * a8s-pg-app
    * Alternatively, the entire demo could be driven by the "assistent" asking the user questions, interactively.


* Don't use the `default` namespace, instead create a demo namespace, e.g. `a8s-demo`.
    * Provision a8s-pg into namespace

* Create binaries in a release matrix, e.g. using Go Release Binaries with Gihub Action Matrix Strategy
    * https://github.com/marketplace/actions/go-release-binaries
    * https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_idstrategymatrix

* Create S3 bucket with configs
    * Alternatively: Install a local storage provider, e.g. minio.
        * Costly dependency: add the local storage provider to the backup agent.
