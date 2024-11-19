---
docs: "DOCS-000"
---

Place the JWT in a file called `license.jwt`. Create a Kubernetes Secret using the contents of the JWT file.

```shell
kubectl create secret generic nplus-license --from-file license.jwt -n nginx-gateway
```

You can now delete the `license.jwt` file.

If you need to update the JWT at any time, update the `license.jwt` field in the Secret using `kubectl edit` and apply the changes.
