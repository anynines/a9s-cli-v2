# a9s CLI V2

# Development

There's a `Makefile` to help building and running the cli during development.

## Build

    make build

The binary can be found in `bin/a9s`.

## Run

    a9s

### Deployment Version

Select a particular release by providing the `--deployment-version` parameter:

    a9s demo a8s-pg --deployment-version v0.3.0

Use

    a9s demo a8s-pg --deployment-version latest

To get the latest, untagged version of the deployment manifests.

### Kubernetes Provider

If you want to select a particular Kubernetes provider:

    a9s demo a8s-pg -p kind 
    a9s demo a8s-pg -p minikube (default)

Follow the instructions to learn about available sub commands.

### Backup Infrastructure Region

    a9s demo a8s-pg --backup-region us-east-1

**Note**: By default, an existing `backup-config.yaml` will be used. Hence, if you intend to change
your backup config, remove the existing `backup-config.yaml`, first:

    rm demo/deploy/a8s/backup-config/backup-store-config.yaml

## Printing the Demo Working Directory

    a9s demo pwd

## Unattended Mode

It is possible to skip all yes-no questions by **enabling the unattended mode** by passing the `-y` or `--yes` flag:

    a9s demo a8s-pg --yes

## Cleaning Up

In order to delete the Demo cluster run:

    a9s demo delete

**Note**: This will not delete config files stored.

Config files are stored in the demo working directory.

They can be removed with:

    rm -rf $( a9s demo pwd )

# Design Principles / Ideals

* The CLI should not need a tight synchronization with product releases.
    * The release of a new a8s Postgres version, for example, should be working with an existing CLI version.

# Backlog

* Remove Kind.

* Sub command to delete all demo resources.
    * Remove cluster        
        * `a9s demo delete`
    * Remove everything (incl. config files)

* Add `-y` / `--yes` flag to `demo a8s` to confirm all yes/no user dialogs with `yes`. This makes it faster when repeating the process several times.
* Make minikube/kind memory configurable
* Make minikube/kind nr of nodes configurable



* Don't use the `default` namespace, instead create a demo namespace, e.g. `a8s-demo`.
    * Provision a8s-pg into namespace
* Question: Should the demo a8s-pg execute the entire demo or just install the operator? Other commands could be: 
    * a8s-pg 
    * a8s-pg-instance 
    * a8s-pg-app
    * Alternatively, the entire demo could be driven by the "assistent" asking the user questions, interactively.

* Create S3 bucket with configs
    * Alternatively: Install a local storage provider, e.g. minio.
        * Costly dependency: add the local storage provider to the backup agent.
