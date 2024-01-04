# Debugging

## Debugging NGF Remotely in Kubernetes

This section will walk you through how to attach an ephemeral [dlv](https://github.com/go-delve/delve) debugger
container to NGF while it's running in Kubernetes. This will allow you to remotely debug NGF running in Kubernetes
using your IDE.

- Create a `kind` cluster:

   ```console
   make create-kind-cluster
   ```

- Set GOARCH environment variable:

  The [Makefile](/Makefile) uses the GOARCH variable to build the binary and container images. The default value of GOARCH is `amd64`.

  If you are deploying NGINX Gateway Fabric on a kind cluster, and the architecture of your machine is not `amd64`, you will want to set the GOARCH variable to the architecture of your local machine. You can find the value of GOARCH by running `go env`. Export the GOARCH variable in your `~/.zshrc` or `~/.bashrc`.

  ```console
  echo "export GOARCH=< Your architecture (e.g. arm64 or amd64) >" >> ~/.bashrc
  source ~/.bashrc
  ```

  or for zsh:

  ```console
  echo "export GOARCH=< Your architecture (e.g. arm64 or amd64) >" >> ~/.zshrc
  source ~/.zshrc
  ```

- Build debug images and install NGF on your kind cluster:

  - **For NGINX OSS:**

    ```console
    make GOARCH=$GOARCH debug-install-local-build
    ```

  - **For NGINX Plus:**

    ```console
    make GOARCH=$GOARCH debug-install-local-build-with-plus
    ```

  > Note: The default value of GOARCH in the [Makefile](/Makefile) is `amd64`. If you try and debug an amd64 container on an ARM machine you will see the following error in the dlv container logs: `could not attach to pid <pid>: function not implemented`.
  > This is a known issue and the only workaround is to create an arm64 image by specifying `GOARCH=arm64` the above commands.
  > For more information, see this [issue](https://github.com/docker/for-mac/issues/5191)

- Start kubectl proxy in the background:

  ```console
  kubectl proxy &
  ```

- Save the NGF Pod name:

  ```console
  POD_NAME=<NGF Pod>
  ```

- Run the following curl command to create an ephemeral debug container:

  ```console
  curl --location --request PATCH 127.0.0.1:8001/api/v1/namespaces/nginx-gateway/pods/$POD_NAME/ephemeralcontainers \
  --header 'Content-Type: application/strategic-merge-patch+json' \
  --data '{
      "spec":
      {
          "ephemeralContainers":
          [
              {
                  "name": "dlv",
                  "command": [
                      "/bin/sh",
                      "-c",
                      "PID=$(pgrep -f /usr/bin/gateway) && dlv attach $PID --headless --listen 127.0.0.1:40000 --api-version=2 --accept-multiclient --only-same-user=false"
                      ],
                  "image": "dlv-debug:edge",
                  "imagePullPolicy": "Never",
                  "targetContainerName": "nginx-gateway",
                  "stdin": true,
                  "tty": true,
                  "securityContext": {
                      "capabilities": {
                          "add": [
                              "SYS_PTRACE"
                          ]
                      },
                      "runAsNonRoot":false
                  }
              }
          ]
      }
  }'
  ```

- Verify that the dlv API server is running:

  ```console
  kubectl logs -n nginx-gateway $POD_NAME -c dlv
  ```

  you should see the following log:

  ```text
  API server listening at: 127.0.0.1:40000
  ```

- Kill the kubectl proxy process:

  ```console
  kill <kubectl proxy PID>
  ```

- Port-forward the dlv API server port on the NGF Pod:

  ```console
  kubectl port-forward -n nginx-gateway $POD_NAME 40000
  ```

- Connect to the remote dlv API server through your IDE:
  - [jetbrains instructions](https://www.jetbrains.com/help/go/attach-to-running-go-processes-with-debugger.html)
  - [vscode instructions](https://github.com/golang/vscode-go/blob/master/docs/debugging.md)

- Debug!
