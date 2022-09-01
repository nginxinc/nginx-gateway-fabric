# Releases

NGINX Kubernetes Gateway uses semantic versioning for its releases. For more information see https://semver.org.

Warning: Major version zero (0.y.z) is reserved for development, anything MAY change at any time. The public API is not stable.

NGINX Kubernetes Gateway, especially in its early stages, will release periodically on a feature set cadence. By feature set, we intend to mean 2-4 features as tracked by the repository's Github milestones. We do this to remain agile and to expose features relatively quickly, but also to minimize the cost of many frequent releases and the associated process overhead.

This means users can expect a release generally every 1 or 2 months, and can also track the progress of a release and its features by viewing our Github milestones. Each issue scheduled for a release will be labelled with "enhancement" (initial feature requests are "proposals", the label "enhancement" will be exchanged when work is committed) and added to the milestone release, milestones will use the form vMAJOR.MINOR. Multiple issues may be written to describe one use-case and each will be organized according to this process. Sub-issues should also link to their parent (the parent issue will list the links via mention notifications).

Bugs and other issue types will use a similar format; once committed to a release they will be assigned to the milestone representing that release.

When all milestone work is complete, we will begin our release process and publish artifacts accordingly.

Example workflow (see issue process section for more detail [TBD]):

1. Discuss open features when planning.

1. Agree and commit to work.

1. Add features to next milestone.

1. Feature "proposal" label removed and "enhancement" label added.

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
