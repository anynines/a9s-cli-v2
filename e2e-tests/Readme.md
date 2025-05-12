# Readme

## Idea

The idea is that Ruby is more user friendly when it comes to writing scripts than Go. Therefore, the test suite of executing the `a9s` executable is written in Ruby and not in Go.

## Assumptions

* The `a9s` cli is installed on your system including all necessary dependencies.
  * Verify prerequisites by using the `a9s` command manually [1].
  * Both `minikube` and `kind` are expected to be available.
* It is assumed that you have already run a demo successfully and hence the workdir is already there including backup store credentials.
  * **Attention: This test suite will interact with your demo system as well as with the backup store bucket (e.g. S3 bucket). DO NOT RUN THIS TESTSUITE IF YOU NEED A WORKING DEMO ENVIRONEMENT!**

## Execute the End-to-End Test Suite

### Cold-Run

1. Install the `a9s` cli and ensure it's in your `$PATH`.
2. Run the `a9s demo a8s` for both `kind` and `minikube`. This ensures that all dependencies are there and creates a local working directory.
3. Create an AWS S3 bucket (one of the e2e-tests uses Minikube and AWS S3 as its backup storage)
4. Set the following ENV vars with your AWS account's credentials:
  `export AWS_ACCESSKEYID=<...>`
  `export AWS_SECRETKEY=<...>`  
  `export ASW_S3_BUCKET_NAME=<...>`
5. In the `e2e-tests` directory run the test suite by executing: `bundle exec rspec`
  * Before the first execution: 
    * [Recommended but optional] Install the Docker Pull Through Registry as this will speed up the test execution reducing the runtime to about 20%.    
    * Ensure that you have installed both Ruby and Bundler.
    * Execute `bundle install`

### Failing Fast

During development it is often helpful to stop the test suite execution after the first failing test. The test environment is then in a suitable state to debug the failing test.

This behavior can be achieved by executing:

    bundle exec rspec --fail-fast

### Executing Test by its Description

    bundle exec rspec --fail-fast -e "verifies the execution of a9s version"

## Test Logs

A log file named `test.log` is created or emptied each time the test suite is executed.

## Docker Pull Through Registry (DPTR)

The idea of the DPTR is to enable container image caching. As the test suite creates fresh Kubernetes clusters all container images are uncached on the per-Node cache of the clusters. Therefore, the major fraction of the test execution time is pulling container images. This can be signficantly reduced when using a DPTR on the level of Docker. As both Kind and Minikube use Docker the caching mechanism will automatically apply to both Kubernetes variants and thus speed up test execution, significantly.

See [1] for more details.

    docker run -d -p 5495:5000 --restart always --name registry registry:2

    docker exec -it registry /bin/sh
    vi /etc/docker/registry/config.yml

Add the following section:

    proxy:
      remoteurl: https://registry-1.docker.io
      username: [username]
      password: [password]

**NOTE**: Don't forget to replace `[username]` and `[password]` with a valid dockerhub username and password.

On MacOS with Docker Desktop edit: `~/.docker/daemon.json` and add:

    "registry-mirrors": ["https://localhost:5495"]

Reload docker.

Run the tests.

## Links

0. https://github.com/anynines/a9s-cli-v2

### Pull Through Registry Links

1. Pull-Through-Registry, https://docs.docker.com/docker-hub/mirror/
2. https://distribution.github.io/distribution/recipes/mirror/
3. https://hub.docker.com/_/registry
4. https://github.com/opencontainers/distribution-spec
5. https://github.com/distribution/distribution
6. https://pkg.go.dev/github.com/distribution/distribution
7. https://distribution.github.io/distribution
