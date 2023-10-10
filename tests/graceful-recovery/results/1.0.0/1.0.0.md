# Results for v1.0.0

## Versions

NGF version:

```text
commit: 72b6c6ef8915c697626eeab88fdb6a3ce15b8da0
date: 2023-10-02T13:13:08Z
version: edge
```

with NGINX:

```text
nginx/1.25.2
built by gcc 12.2.1 20220924 (Alpine 12.2.1_git20220924-r10)
OS: Linux 5.15.49-linuxkit-pr
```


Kubernetes:

```text
Server Version: version.Info{Major:"1", Minor:"28",
GitVersion:"v1.28.0",
GitCommit:"855e7c48de7388eb330da0f8d9d2394ee818fb8d",
GitTreeState:"clean", BuildDate:"2023-08-15T21:26:40Z",
GoVersion:"go1.20.7", Compiler:"gc",
Platform:"linux/arm64"}
```


### Restart nginx-gateway container
Passes test with no errors.

### Restart NGINX container
Passes test with no errors.

### Restart Node with draining
Passes test with no errors.

### Restart Node without draining
Does not work correctly the majority of times and errors after running `docker restart kind-control-plane`.
NGF Pod is not able to recover as the NGINX container logs show this error:
`bind() to unix:/var/run/nginx/nginx-status.sock failed (98: Address in use)`.

Issue Filed: https://github.com/nginxinc/nginx-gateway-fabric/issues/1108
