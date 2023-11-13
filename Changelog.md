# V0.3.0

* New command structure

    * `a9s demo a8s-pg`: Executes the demo and is preferred over `a9s a8s-pg-demo`. This fascilitates adding more demos in future releases.
    * `a9s demo pwd`: Prints the current demo working directory.

# v0.2.1
* BUGFIX for issue where the a8s-pg-demo was crashing as kubeconfig flag was already defined.

# v0.2.0

* FEATURE: Now waits for a8s-system pods to become ready, so that a user knows when the system is ready to create service instances.

# v0.1.0

* Initial version