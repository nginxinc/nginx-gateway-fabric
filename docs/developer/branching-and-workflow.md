# Branching and Workflow

NKG is an open source and public repository; our goal is for a spartan and easily manageable set of branches. Ideally,
our branches are created minimally; there will be a main branch and a handful of release and other long-term feature
branches.

Internal developers and external contributors will follow a fork and merge process. Each contributor should fork the
repo to their own space; branch, experiment, develop and merge from their space into NKG’s main branch. NKG branches are
long-lived, developer branches are ephemeral and unconstrained outside the main repository.

Developer Alice,

- Forks `github.com/nginxinc/nginx-kubernetes-gateway` → `github.com/alice/nginx-kubernetes-gateway`
- Alice develops a feature or bugfix - using as many ephemeral branches as she needs.
- Alice creates a
  PR `github.com/alice/nginx-kubernetes-gateway:feature/some-feature` → `github.com/nginxinc/nginx-kubernetes-gateway:main`

## Branch Naming Conventions

To maintain consistency and facilitate proper labeling of pull requests (PRs), we follow specific branch naming
conventions. Each branch should contain a prefix that accurately describes the purpose of the PR. The prefixes align
with the labels defined in the [labeler](../../.github/labeler.yml) file, which are used to create release notes.

For example, if you are working on a bug fix, name your branch starting with `bug/` or `fix/`, followed by a descriptive
name of the bug you are fixing.

To ensure correct labeling of your PRs, please use the appropriate prefix from the predefined list when naming your
branches. This practice helps maintain consistent labeling and allows for the automated generation of accurate release
notes.

For a comprehensive list of labels and prefixes, please refer to the [labeler](../../.github/labeler.yml) file.
