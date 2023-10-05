# Test Results

## Testing when nginx-gateway container restarts
Passes test with no errors.

## Testing when nginx container restarts
Passes test with no errors.

## Testing when the NGF Pod restarts through a docker container restart with a graceful exit on node
Passes test with no errors.

## Testing when the NGF Pod restarts through a docker container restart without cleaning up the node prior
Does not work correctly the majority of times and errors after running `docker restart kind-control-plane`.
NGF Pod is not able to recover as the nginx container logs show this error:
`bind() to unix:/var/run/nginx/nginx-status.sock failed (98: Address in use)`.
