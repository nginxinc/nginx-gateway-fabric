# Releases

NGINX Kubernetes Gateway uses semantic versioning for its releases. For more information see https://semver.org.

Warning: Major version zero (0.y.z) is reserved for development, anything MAY change at any time. The public API is not stable.

NGINX Kubernetes Gateway, especially in its early stages, will release periodically on a feature set cadence. By feature set, we intend to mean 2-4 features as tracked by the repository's Github milestones. We do this to remain agile and to expose features relatively quickly, but also to minimize the cost of many frequent releases and the associated process overhead.

This means users can expect a release generally every 1 or 2 months, and can also track the progress of a release and its features by viewing our Github milestones. Each issue scheduled for a release will be labelled with "enhancement" (initial feature requests are "proposals", the label "enhancement" will be exchanged when work is committed) and added to the milestone release, milestones will use the form vMAJOR.MINOR.PATCH. Multiple issues may be written to describe one use-case and each will be organized according to this process. Sub-issues should also link to their parent (the parent issue will list the links via mention notifications).

Bugs and other issue types will use a similar format; once committed to a release they will be assigned to the milestone representing that release.

When all milestone work is complete, we will begin our release process and publish artifacts accordingly.

Example workflow (see issue process section for more detail [TBD]):

1. Discuss open features when planning.

1. Agree and commit to work.

1. Add features to next milestone, for example, v0.2.0.

1. Feature "proposal" label removed and "enhancement" label added.

### Steps to create a release.

NGINX Kubernetes Gateway is a trunk based development project; generally following a modified "Branch for release" pattern. Branches will be created as late as possible, often only when a patch is required and the release will first diverge from the main trunk. This results in tags created for planned releases, but branches for patch releases.

#### Planned releases

1. When all committed items are complete within a milestone a release will be created.

1. Trunk merges are ceased for testing. Parallel development is free to continue on development branches (branching can occur here, but we will prefer tagging on trunk for planned releases).

1. Trunk is tested using the "edge" container builds.

  If tests fail:
  
    - Fix for issue is created.

    - Open PR against main (trunk) branch.

    - New commit and merge will create a new "edge" container.

    - Repeat.

1. Once tests have passed create a tag (in the form vMAJOR.MINOR.PATCH) on the main (trunk) commit SHA. The docker image will automatically be pushed to `ghcr.io/nginxinc/nginx-kubernetes-gateway:MAJOR.MINOR.PATCH` with the docker tag reflecting the planned release version.

1. Update the changelog with the changes added in the release. They can be found in the github release notes that was generated from the release branch.

#### Patch releases

1. If no release branch exists.

  1. Create a release branch from the planned release tag in main (for example, v0.2.0), use the naming format: release-MAJOR.MINOR.

1. Reproduce and fix the bug in main first. This process helps ensure all issues remain fixed in trunk and aren't missed.

  If the tests fail:

    - Create a fix for the error.

    - Open a PR with the fix against the main branch.

    - New commit and merge will create a new "edge" container.

    - Repeat.

1. Cherry-pick the commit from main to the release branch. Use the `-x` argument to preserve the cherry pick's origin commit.

1. Test the release branch.

  If the tests fail (this may occur because the trunk has diverged or a poor conflict resolution):

    - Create a fix for the error.

    - Open a PR with the fix against the release branch.

    - Repeat.

1. Once tests have passed create a tag (in the form vMAJOR.MINOR.PATCH) on the release branch commit SHA. The docker image will automatically be pushed to `ghcr.io/nginxinc/nginx-kubernetes-gateway:MAJOR.MINOR.PATCH` with the docker tag reflecting the planned release version.

1. Update the changelog with the changes added in the release. They can be found in the github release notes that was generated from the release branch.
