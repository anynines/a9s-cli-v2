# a9s CLI

anynines provides a command line tool called `a9s` to fasciliate application development, devops tasks and interact with selected anynines products.

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

# Next Steps
Please refer to the [a9s CLI](https://docs.a9s-cli.anynines.com) documentation for detailed instructions as well as a [hands-on tutorial](https://docs.a9s-cli.anynines.com/docs/hands-on-tutorials/).

# Links
* a9s CLI Documentation, https://docs.a9s-cli.anynines.com/
* a9s CLI hands-on tutorial, https://docs.a9s-cli.anynines.com/docs/hands-on-tutorials/