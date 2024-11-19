---
docs: "DOCS-000"
---

{{< note >}} If you would rather pull the NGINX Plus image and push to a private registry, you can skip this specific step and instead follow [this step]({{<relref "installation/nginx-plus-jwt.md#pulling-an-image-for-local-use">}}). {{< /note >}}

If the `nginx-gateway` namespace does not yet exist, create it:

```shell
kubectl create namespace nginx-gateway
```

Create a Kubernetes `docker-registry` secret type using the contents of the JWT as the username and `none` for password (as the password is not used).  The name of the docker server is `private-registry.nginx.com`.

```shell
kubectl create secret docker-registry nginx-plus-registry-secret --docker-server=private-registry.nginx.com --docker-username=<JWT Token> --docker-password=none -n nginx-gateway
```

It is important that the `--docker-username=<JWT Token>` contains the contents of the token and is not pointing to the token itself. When you copy the contents of the JWT, ensure there are no additional characters such as extra whitespaces. This can invalidate the token, causing 401 errors when trying to authenticate to the registry.
