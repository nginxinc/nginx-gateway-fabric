# Releases

NGINX Kubernetes Gateway uses semantic versioning for its releases. For more information see https://semver.org.

Warning: Major version zero (0.y.z) is reserved for development, anything MAY change at any time. The public API is not stable.

NGINX Kubernetes Gateway, especially in its early stages, will release periodically on a feature set cadence. By feature set, we intend to mean 2 or 3 features as tracked by the repository's Github issues. We do this to remain agile and to expose features relatively quickly, but also to minimize the cost of many frequent releases and the associated process overhead.

This means users can expect a release generally every 1 or 2 months, and can also track the progress of a release and its features by viewing our Github issues. Each issue scheduled for a release will be labelled with "proposal" and its scheduled release, the label will use the form vMAJOR.MINOR. Multiple issues may be written to describe one use-case and each will be labelled according to this process. Sub-issues should also link to their parent (the parent issue will list the links via mention notifications).

Bugs and other issue types will use a similar format; once committed to a release they will be assigned the label of that release.

When all labelled work is complete, we will begin our release process and publish artifacts accordingly.

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
