# Implementing a Feature

This section provides guidance on implementing new features in the project. Follow the recommended steps and best
practices to ensure a successful feature development process.

> **Note**
>
> If you’d like to implement a new feature, please open a discussion about the feature before creating an issue or opening a PR.

1. **Assign yourself to the GitHub issue for the feature**: Each feature must have a corresponding GitHub issue to track
   its progress.
2. **Post any questions or comments about the feature on the corresponding issue**: This allows for better tracking and
   visibility. If any discussions regarding the issue occur outside the issue thread, provide a summary of the
   conversation as a comment on the issue itself. This ensures that all relevant information and discussions are
   consolidated in one place for easy reference.
3. **Fork the repo**: NKG follows a fork workflow, which you can learn more about in
   the [branching and workflow](branching-and-workflow.md) documentation.
4. **Branch**: Create a branch following the [naming conventions](branching-and-workflow.md#branch-naming-conventions).
5. **Make changes**: Make the necessary changes for the feature.
6. **Consider opening a draft PR**: If your feature involves substantial architecture changes, or you would like to
   receive early feedback, consider opening a draft PR and requesting reviews from the maintainers. Draft PRs are an
   effective means to solicit feedback and ensure that major architectural changes align with the project's goals and
   standards.
7. **Add or update unit tests** All features **must** be accompanied by unit tests that verify the functionality. Make
   sure to thoroughly test the different scenarios and edge cases related to the feature to ensure robustness and
   reliability. Additionally, open the code coverage report to ensure that the code you added has sufficient test
   coverage. For instructions on writing and running unit tests, refer to the [testing](testing.md#unit-test-guidelines)
   documentation.
8. **Manually verify your changes**: Refer to the [manual testing](testing.md#manual-testing) section of the testing
   documentation for instructions on how to manually test your changes.
9. **Update any relevant documentation**: Here are some guidelines for updating documentation:
    - **Gateway API Feature**: If you are implementing a Gateway API feature, make sure to update
      the [Gateway API Compatibility](../gateway-api-compatibility.md) documentation.
    - **New Use Case:** If your feature introduces a new use case, add an example of how to use it in
      the [examples](../../examples) directory. This example will help users understand how to leverage the new feature
      in their own deployments.
    - **Installation Changes**: If your feature involves changes to the installation process of NKG, update
      the [installation](../installation.md) documentation.
    - **Command-line Changes**: If your feature introduces or changes a command-line flag or subcommand, update
      the [cli help](../cli-help.md) documentation.
    - **Other Documentation Updates**: For any other changes that affect the behavior, usage, or configuration of NKG,
      review the existing documentation and update it as necessary. Ensure that the documentation remains accurate and
      up to date with the latest changes.
10. **Lint code**: See the [run the linter](quickstart.md#run-the-linter) section of the quickstart guide for
    instructions.
11. **Open PR**: Open a PR targeting the `main` branch of
    the [nginx-kubernetes-gateway](https://github.com/nginxinc/nginx-kubernetes-gateway/tree/main) repository. The
    entire `nginx-kubernetes-gateway` group will be automatically requested for review. If you have a specific or
    different reviewer in mind, you can request them as well. Refer to the [pull request](pull-request.md) documentation
    for expectations and guidelines.
12. **Obtain the necessary approvals**: Work with code reviewers to maintain the required number of approvals.
13. **Squash and merge**: Squash your commits locally, or use the GitHub UI to squash and merge. Only one commit per PR
    should be merged.

## Fixing a bug

When fixing a bug, follow the same process as [implementing a feature](#implementing-a-feature) with one additional
requirement:

All bug fixes should be reproduced with a unit test before submitting any code. Once the bug is reproduced in a unit
test, make the necessary code changes to address the issue and ensure that the unit test passes successfully. This
systematic approach helps ensure that the bug is properly understood, effectively resolved, and prevents regression.
