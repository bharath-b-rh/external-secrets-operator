# Scope constraints (copy to openspec/changes/<name>/inputs/scope-constraints.md)

## Jira

- **Only** `inputs/jira-spec.md` (pilot template: `docs/openspec-eval-pilot-jira.md`)
- Forbidden: Jira MCP, linked issues, epic, subtasks, comments, attachments

## Repository

- **Only** files under the working folder at current `git rev-parse HEAD`
- Forbidden: `git fetch`, `git checkout main`, diff vs `main`/`origin/*`, GitHub API, other branches, remote clone

## If information is missing

- Mark `[NEEDS CLARIFICATION]` or state an assumption in the artifact
- Do NOT pull missing context from Jira links or other branches
