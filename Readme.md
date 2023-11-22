# a9s CLI V2

# Development

There's a `Makefile` to help building and running the cli during development.

## Build

    make build

The binary can be found in `bin/a9s`.

## Run

    bin/a9s

Follow the instructions to learn about available sub commands.

# Backlog

* Use Minikube instead of kind

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

* Sub command to delete all demo resources.
    * Remove cluster
    * Remove everything (incl. config files)