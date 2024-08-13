# a9s CLI

anynines provides a command line tool called `a9s` to fasciliate application development, devops tasks and interact with selected anynines products.

### In Action

[![asciicast](https://asciinema.org/a/669151.svg)](https://asciinema.org/a/669151)

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

### `klutch` Stack
* Install a local Klutch central management cluster using `kind`
* Install Crossplane and the a8s stack on the central management cluster
* Bind resources from a consumer cluster to the management cluster

# Next Steps
Please refer to the [a9s CLI](https://docs.a9s-cli.anynines.com) documentation for detailed instructions as well as a [hands-on tutorial](https://docs.a9s-cli.anynines.com/docs/hands-on-tutorials/).
