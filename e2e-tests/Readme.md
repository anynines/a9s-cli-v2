# Readme

## Idea

The idea is that Ruby is more user friendly when it comes to writing scripts than Go. Therefore, the test suite of executing the `a9s` executable is written in Ruby and not in Go.

## Assumptions

* The `a9s` cli is installed on your system including all necessary dependencies.
  * Verify prerequisites by using the `a9s` command manually [1].
  * Both `minikube` and `kind` are expected to be available.
* It is assumed that you have already run a demo successfully and hence the workdir is already there including backup store credentials.
  * **Attention: This test suite will interact with your demo system as well as with the backup store bucket (e.g. S3 bucket). DO NOT RUN THIS TESTSUITE IF YOU NEED A WORKING DEMO ENVIRONEMENT!**

## Execute the Test Suite

1. Install the `a9s` cli and ensure it's in your `$PATH`.
2. Run the `a9s demo a8s` for both `kind` and `minikube`. This ensures that all dependencies are there and creates a local working directory.
3. Run this test suite: `bundle exec rspec`

## Test Logs

A log file named `test.log` is created or emptied each time the test suite is executed.

## Docker Pull Through Registry

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
