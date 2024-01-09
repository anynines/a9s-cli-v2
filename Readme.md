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

### Creating a Custom Service Instance

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

## Applying a SQL File to a Service Instance

Uploading a SQL file, executing it using `psql` and deleting the file can be done with:

    a9s pg apply --file /path/to/sql/file --instance-name sample-pg-cluster

Use `--yes` to skip the confirmation prompt.

    a9s pg apply --file /path/to/sql/file --instance-name sample-pg-cluster --yes

Use `--no-delete` to leave the file in the pod:

    a9s pg apply --file /path/to/sql/file --instance-name sample-pg-cluster --no-delete

## Creating a Backup of a Service Instance

    a9s create pg backup --name sample-pg-cluster-backup-1 -i sample-pg-cluster-1

## Restoring a Backup of Service Instance

    a9s create pg restore --name sample-pg-cluster-restore-1 -b sample-pg-cluster-backup-1 -i sample-pg-cluster-1

## Cleaning Up

In order to delete the Demo cluster run:

    a9s delete demo a8s

**Note**: This will not delete config files stored.

Config files are stored in the demo working directory.

They can be removed with:

    rm -rf $( a9s demo pwd )

# Testing the CLI

The state of unit tests is currently very poor.

End-to-end testing can be done using the external Ruby/RSpec test suite located at: https://github.com/anynines/a9s-cli-v2-tests

# Design Principles / Ideals
* The CLI acts like a personal assistent who knows the a9s products and helps to use them more easily.
    * The CLI helps with installing a demo environment
    * The CLI helps with writing YAML manifests, e.g. so that users do not have to lookup attributes in the documentation.
* The CLI should not need a tight synchronization with product releases.
    * The release of a new a8s Postgres version, for example, should be working with an existing CLI version.

# Known Issues

* Creating a backup for non-existing service instances falsely suggests that the backup has been successful.
* Deletion of backups with `kubectl delete backup ...` get stuck and the deletion doesn't succeed.
* When applying a sql file to an a8s Postgres database using `a9s pg apply --file` ensure that there is no change of the primary pod for clustered instances as otherwise the file might be copied to the wrong pod. There's a slight delay between determining the primary pod and uploading the file to it. 