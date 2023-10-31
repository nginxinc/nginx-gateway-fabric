# Branching and Workflow

NGF is an open source and public repository; our goal is to keep the number of branches in the repository to a minimum:
the main branch, release branches and long-term feature branches.

Internal developers and external contributors will follow a fork and merge process. Each contributor should fork the
repo to their own space; branch, experiment, develop and prepare a pull request (PR) to merge their work into NGF’s main
branch. This way ephemeral developer branches will remain outside the main repository.

Below is an example of following the merge and fork process. Developer Alice:

- Forks `github.com/nginxinc/nginx-gateway-fabric` → `github.com/<alice-user-id>/nginx-gateway-fabric`
- Adds upstream:

  ```shell
  git remote add upstream git@github.com:nginxinc/nginx-gateway-fabric.git
  ```

- Alice lists all of her configured remotes:

  ```shell
  git remote -v
  ```

  Which shows the following:

  ```text
  origin	git@github.com:<alice-user-id>/nginx-gateway-fabric.git (fetch)
  origin	git@github.com:<alice-user-id>/nginx-gateway-fabric.git (push)
  upstream	git@github.com:nginxinc/nginx-gateway-fabric.git (fetch)
  upstream	git@github.com:nginxinc/nginx-gateway-fabric.git (push)
  ```

- Alice develops a feature or bugfix - using as many ephemeral branches as she needs.
- Alice creates a
  PR `github.com/<alice-user-id>/nginx-gateway-fabric:feature/some-feature` → `github.com/nginxinc/nginx-gateway-fabric:main`
- Alice keeps her fork up to date by running:

  ```shell
  git pull upstream main
  ```

  This will sync her local main branch with the main branch of the project's repo.

## Branch Naming Conventions

To maintain consistency and facilitate proper labeling of pull requests (PRs), we follow specific branch naming
conventions. Each branch should contain a prefix that accurately describes the purpose of the PR. The prefixes align
with the labels defined in the [labeler](/.github/labeler.yml) file, which are used to create release notes.

For example, if you are working on a bug fix, name your branch starting with `bug/` or `fix/`, followed by a descriptive
name of the bug you are fixing.

To ensure correct labeling of your PRs, please use the appropriate prefix from the predefined list when naming your
branches. This practice helps maintain consistent labeling and allows for the automated generation of accurate release
notes.

For a comprehensive list of labels and prefixes, please refer to the [labeler](/.github/labeler.yml) file.
