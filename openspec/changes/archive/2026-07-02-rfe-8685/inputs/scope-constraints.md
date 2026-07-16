# Scope constraints

## Jira
- **Only** `inputs/jira-spec.md`
- Forbidden: Jira MCP, linked issues, epic, subtasks, comments, remote issue fetch

## Repository
- **Only** files under `working_folder_path` at current `git rev-parse HEAD`
- Forbidden: `git fetch`, `git checkout main`, diff vs `main`/`origin/*`, GitHub API, other branches

## If information is missing
- Mark `[NEEDS CLARIFICATION]` or state an assumption in the artifact
- Do NOT pull missing context from Jira links or other branches
