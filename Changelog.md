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