# Pull Requests

## Submitter Guidelines

- Fill in [our pull request template](../../.github/PULL_REQUEST_TEMPLATE.md). In addition, write a clear and concise PR
  description that helps reviewers understand the purpose and impact of your changes. Start with a brief overview of the
  problem or feature being addressed. Then, explain the approach you took to implement the solution, highlighting any
  significant design decisions or considerations. Finally, mention any specific areas where you would like reviewers to
  focus their attention or provide specific feedback. A well-written PR description sets the stage for effective code
  review and facilitates a smoother review process.
- Include the issue number in the PR description to automatically close the issue when the PR mergers.
  See [Closing Issues via Pull Requests](https://github.blog/2013-05-14-closing-issues-via-pull-requests/) for details.
- For significant changes, break your changes into a logical series of smaller commits. By approaching the changes
  incrementally, it becomes easier to comprehend and review the code.
- Use your best judgement when resolving comments. Simple, unambiguous comments should be resolved by the submitter,
  this prioritizes speed and efficacy. For larger changes, for example, when updating logic and design, the resolution
  should be left for the reviewer’s approval or continued discussion.
- If a discussion grows too long - more than a handful (3 - 5) back and forth responses, consider moving the discussion
  to a more real-time platform: Slack, Zoom, Phone, In-person. Do this to clear misunderstandings or whiteboard
  improvements. Whenever the conversation has moved to a different platform and found a conclusion, the decision and
  resolution MUST be updated in the review, for instance,

  ```text
  Comment:

  Spoke offline, decided to move A to package B. Rewrite tests and API docs.
  ```

- When pushing code review changes, it's important to provide clear information to the reviewer. Here are a couple of
  guidelines to follow:
    - Commit each code review change individually.
    - Use descriptive commit messages that accurately describe the change made, making it easy for the reviewer to
      understand the purpose of the change.
- When approved, and all review comments have been resolved; squash commits to a single commit. Follow
  the [Commit Message Format](#commit-message-format) guidelines when writing the final commit.

## Reviewer Guidelines

- As a reviewer, you bear the ultimate responsibility for the resolution of the comment. If not resolved, please
  follow-up in a timely manner with acceptance of the solution and explanation.
- Do your best to review in a timely manner. However, code reviews are time-consuming, maximize their benefit by
  focusing on what’s highest value. This code review pyramid outlines a reasonable shorthand:
  ![Code Review Pyramid](code-review-pyramid.jpeg)
- Always review for: design, API semantics, [DRY](https://en.wikipedia.org/wiki/Don%27t_repeat_yourself) and
  maintainability practices, bugs and quality, efficiency and correctness.
- Code review checklist:
    - Do any changes need to made in documentation?
        - Do public docs need updating?
        - Have internal workflow and informational docs been impacted?
    - Are there unit tests for the change?
        - If the change is a bug fix, has a unit test been added or modified that catches the bug? All bug fixes should
          be reproduced with a unit test before submitting any code. Then the code should be fixed so that the unit test
          passes.
        - If the change is a feature or new functionality. The expectation is to include a reasonable set of unit tests
          which exercise positive and negative cases.
    - Is this change backwards compatible? Is it causing breaking behavior?
    - Are there any [Comment Tags](#comment-tags) in the change?

## Commit Message Format

The commit message for your final commit should use a consistent format:

```text
<One line summary of the change>

Problem: Brief description of the problem. The “why”.

Solution: Detailed description of the solution. Try to refrain from pseudo-code. Describe some of the analysis, thought
process, and outline what was done. Think of a newcomer coming across this two years hence.

Testing (optional): Description of tests that were run to exercise the solutions (unit tests, system tests,
reproductions, etc.)

```

Here's an example:

```text
Make event stream handle large events

Problem: The event stream would choke on events larger than 1K. This caused events to be silently dropped, and affected
all downstream consumers.

Solution: Removed upper limit on event size. Also, added an error handler to warn when events are being dropped for any
reason.

Added unit tests for events up to 1MB.
```

> **Note**
> Do not put any customer identifying information in message. Say “a customer found…”, NOT “ACME Corp found…”. If customer generated data is included it must be redacted for any PII (or identifying information). This includes IPs, ports, names of applications and objects, and URL paths. If customers volunteered this information, do not propagate it externally.

For additional help on writing good commit messages, see this [article](https://cbea.ms/git-commit/).

# Comment Tags

Common comment tags are FIXMEs and TODOs.

A TODO is a developer note-to-self. TODOs _MUST_ be fixed before merging.

A FIXME is for things that should be fixed in the future. FIXMEs can be merged, but they _MUST_ include your username and
an link to the issue. For example:

```go
// FIXME(username): This is currently a hack to work around known issue X. 
// Issue <LINK TO GITHUB ISSUE> will remove need for this work around.
```
