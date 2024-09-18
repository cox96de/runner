# About

This project is a CI/CD pipeline engine similar to Jenkins. `Runner` is a self-contained, open-source automation server
designed to facilitate various tasks, including building, testing, and deploying software.

`Runner` supports multiple execution environments and offers the following key features:

- Execute build and test jobs on the host machine.

- Run build and test jobs on a Kubernetes pod, which Runner manages by creating and destroying it as needed.

- Execute build and test jobs on Windows and Linux QEMU virtual machines, also managed by Runner.

- Support for multiple operating systems and architectures, including macOS, Linux, and Windows.

- Support for Directed Acyclic Graph (DAG) pipelines.

# Quick start

The `Runner` contains two components: `Runner Server` and `Runner Agent`.

## Install `Runner Server`

`Runner Server` is the api server for `Runner`.

Install `Runner Server` by docker

```bash
docker run -p 8080:8080 cox96de/runner-server:latest
```

## Install `Runner Agent`

`Runner Agent` is the worker for `Runner`. It deploys on the worker machine.

Install `Runner Agent` by docker

```bash
docker run cox96de/runner-agent-debian:latest --engine shell --server_url http://{your_server_ip}:8080
```

Notice: replace `{your_server_ip}` with your server ip.

## Run a job

`Runner` provides a simple CLI to easy run a job.

```bash
docker run cox96de/runner-simplecli --server http://{your_server_ip}:8080 /examples/compile_redis.yaml
```
