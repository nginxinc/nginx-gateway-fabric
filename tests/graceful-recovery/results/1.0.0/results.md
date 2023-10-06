# Test Results

## Version 1.0.0

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
