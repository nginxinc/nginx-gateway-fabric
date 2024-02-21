---
docs: "DOCS-1440"
---

Set `terminationGracePeriodSeconds` to a value that is equal to or greater than the `sleep` duration specified in the `preStop` hook (default is `30`). This setting prevents Kubernetes from terminating the pod before before the `preStop` hook has completed running.

   ```yaml
   terminationGracePeriodSeconds: 50
   ```
