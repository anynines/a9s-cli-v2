# v0.10.0

* Change: `a9s delete pg restore` now verifies whether service instance exstist and fails with a non-zero return code if the service instance doesn't exist.
* Change: `a9s delete pg backup` now verifies whether service instance exstist and fails with a non-zero return code if the service instance doesn't exist.
* Change: `a9s delete pg instance` now verifies whether service instance exstist and warns if not existing with return code `0` as the desired state is that the instance shall not exist.

# v0.9.0
* Change: `--yes` is now a global flag and available to all commands.
* Change: `--verbose` or `-v` is now a global flag and available to all commands. Standard output is now less verbose.
* Feature: `a9s create pg instance` to generate a YAML manifest given the params: `--namespace`, `--api-version`, `--name`, `--replicas`, `--volume-size`, `--service-version`, `--requests-cpu` and `--limits-memory`
* Feature: `a9s create pg backup` to generate a backup YAML manifest, execute the backup and wait for it to complete.
* Feature: `a9s create pg restore` to generate a restore YAML manifest, execute the restore and wait for it to complete.
* Feature: Add `--no-apply` flag allow the generation of YAML manifests without applying them.
* BUGFIX: When creating service instance YAML manifests, the namespace of the service instance is now set, correctly.
* BUGFIX: Params for creating pg instances do now belong to the `a9s create pg instance` command instead of `a9s create pg`.
* BUGFIX: The `backup-provider` param in `a9s create demo a8s` is now correctly set instead of being falsely assigned to the `backup-bucket` parameter.
* BUGFIX: executing a9s from an arbitrary file should writeYAML files to the working directory not relative to the exeuction folder of the a9s binary.
* BUGFIX: The filename of a backup manifest should be correct but is: usermanifests/a8s-pg-backup-a8s-pg-backup.yaml
* BUGFIX: Creating a service instance named `solo` with a single replica should not print output containing the name `clustered-0` due to assuming any system to consist of 3 replicas.
* Default change: Makes `eu-central-1` the default infrastructure region.
* Removes Docker as a necessary prerequisite as not all Kubernetes providers mandatorily need Docker
* Testing: Created a Ruby/RSpec test suite to run the demo automatically for both `kind` and `minikube`. See: https://github.com/anynines/a9s-cli-v2-tests

# v0.8.0

* The `a8s-deployment` repository is now cloned to {workdir}/a8s-deployment and is not at the {workdir} root anymore.
    * This allows cloning additional repositories as the same level.
* Now also the `a8s-demo` repository with its demo application manifests is cloned into {workdir}.
    * In theory the `a8s-demo` repo contains `a8s-deployment` as a submodule. However, using two separate `git clone` operations provides a more granular control over which version of each repository to checkout.

* Create Postgres service instances with `a9s create pg instance`
* Delete Postgres service instances with `a9s delete pg instance`
* Create an a8s Data Service demo environment with `a9s create demo a8s` which used to be `a9s demo a8s-pg`.
* Delete an a8s Data Service demo environment with `a9s delete demo a8s` which used to be `a9s demo delete`.

# v0.7.0

* Introduces a multi-k8s-provider support with implementations for both `minikube` and `kind`.
* The Kubernetes Node memory is now `4gb` instead of `8gb` per default.
* The number of Kubernetes Nodes is now `3` instead of `4` per default.

# v0.6.0 

* Adds a parameter to skip the verification of prerequisites `a9s demo a8s-pg --no-prechecks`
* Adds parameters to set the nr of Kubernetes nodes as well as the cluster memory: `a9s demo a8s-pg --cluster-nr-of-nodes 1 --cluster-memory 12gb`
* Adds parameter to select the version of the a8s-deployment manifests. See Readme.md for more details.

# v0.5.1

* Fixes bug in CLI params `--backup-bucket`, `--backup-region` and `backup-provider`. Backup region has also been used as a backup name and backup-provider was fixed to "AWS" instead of using the param.

# v0.5.0

* Removes support for kind to focus exclusively on minikube for now.
* Adds a command for deleting the demo kubernetes cluster `a9s demo delete`.
* Adds unattended mode allowing to run demos faster by skipping yes-no-questions: `a9s demo a8s-pg --yes`

# v0.4.1
* Adds missing `Makefile`
* Moves the `-p` or `--provider` flag to the subcommand `a9s demo a8s-pg`.
* The settings for the a8s Backup subsystem can now be configured:
    * Settings:
        * Infrastructure provider, e.g. `"AWS"`
        * Bucket name, e.g. `"a8s-backups"`
        * Infrastructure region, e.g. `"us-east-1"`
    * See `a9s demo a8s-pg --help` for details.

# v0.4.0

* Minikube support added.
* Minikube is the new default instead of kind.
* Minikube creates a 4 node cluster while kind creates a single node cluster.
* Minikube cluster is configured to use 8GB of memory.

# v0.3.0

* New command structure

    * `a9s demo a8s-pg`: Executes the demo and is preferred over `a9s a8s-pg-demo`. This fascilitates adding more demos in future releases.
    * `a9s demo pwd`: Prints the current demo working directory.
* CHORE: Added Makefile.
* CHORE: Renamed modules.


# v0.2.1
* BUGFIX for issue where the a8s-pg-demo was crashing as kubeconfig flag was already defined.

# v0.2.0

* FEATURE: Now waits for a8s-system pods to become ready, so that a user knows when the system is ready to create service instances.

# v0.1.0

* Initial version