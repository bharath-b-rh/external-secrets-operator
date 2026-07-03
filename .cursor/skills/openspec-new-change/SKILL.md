---
name: openspec-new-change
description: Start openspec-agile-workflow change from Jira ticket. Use for /opsx-new.
license: MIT
compatibility: Requires openspec CLI.
metadata:
  author: openspec
  version: "1.1"
---

Jira key required at `/opsx-new`. Write `inputs/jira.yaml`. Obtain spec via Jira MCP or user paste into `inputs/jira-spec.md`. Do not create artifacts. Next: `/opsx-continue`.

Syntax: `/opsx-new PROJ-123` or `/opsx-new PROJ-123 change-name`

Repo URL optional now; required before repo-assessment stage.
