# Release Process for NGINX Kubernetes Gateway

NGINX Kubernetes Gateway uses semantic versioning for its releases. For more information see https://semver.org.

Warning: Major version zero (0.y.z) is reserved for development, anything MAY change at any time. The public API is not stable.

### Steps to create a release.

1. Create a release branch from main, use the naming format: release-MAJOR.MINOR.

2. Create a release candidate tag, use the naming format: vMAJOR.MINOR.PATCH-rc.N (N must start from 1 and monotonically increase with each release candidate).

3. Test the release candidate.

    If the tests fail

    - Create a fix for the error.

    - Open a PR with the fix against the main branch.

    - Once approved and merged, cherry-pick the commit into the release branch.

    - Create a new release candidate tag, increment the release candidate number by 1.

4. Iterate over the process in step 3 until all the tests pass on the release candidate tag then create the final release tag from the release branch in the format vMAJOR.MINOR.PATCH.  The docker image will automatically be pushed to ghcr.io/nginxinc/nginx-kubernetes-gateway:MAJOR.MINOR.PATCH with the release tag as the docker tag.

5. Update the changelog with the changes added in the release.  They can be found in the github release notes that was generated from the release branch.
