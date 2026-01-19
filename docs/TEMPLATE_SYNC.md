# Template Synchronization Guide

This guide explains how to pull updates from the upstream template (`go-microservice-template`) into projects that were created from it.

The approach is git-native: add the template as a remote, fetch updates, and merge them into your repository. A helper script and Makefile targets are provided to streamline the process.

## Quick Commands

```bash
make template-setup   # one-time setup: add remote, fetch, create .template-version
make template-status  # show sync status and latest template tags
make template-diff    # show diff vs template/main (or tag)
make template-sync    # merge template/main into current branch
```

### Using a specific template tag
```bash
make template-diff v1.2.0
make template-sync v1.2.0
```

## Files Added for Sync
- `scripts/template-sync.sh` — automation for setup/status/diff/sync
- `.template-version` — tracks last synced template ref (created in downstream repos)
- Makefile targets: `template-setup`, `template-status`, `template-fetch`, `template-diff`, `template-sync`

## One-Time Setup (downstream repositories)
1) Ensure a clean working tree (no uncommitted changes).
2) Run:
```bash
make template-setup
```
This will:
- Add remote `template` -> `https://github.com/andskur/go-microservice-template.git`
- Fetch the template
- Create `.template-version` with initial metadata

## Checking Status and Differences
- Show current sync state and recent template tags:
  ```bash
  make template-status
  ```
- Show summary diff vs template/main:
  ```bash
  make template-diff
  ```
- Diff against a specific template tag:
  ```bash
  make template-diff v1.2.0
  ```

For detailed diffs of specific files, use standard git:
```bash
git diff template/main -- path/to/file
```

## Syncing Template Updates
1) Ensure a clean working tree.
2) Fetch and merge:
```bash
make template-fetch
make template-sync           # merges template/main
make template-sync v1.2.0    # merges a specific tag
```
3) If conflicts occur, resolve them manually, then re-run `make template-diff` or `git status` to verify.
4) On successful merge, `.template-version` is updated automatically with the merged ref and date.
5) Run tests: `make test` (and optionally `make build`).
6) Commit with a clear message, e.g., `chore: sync from template v1.2.0`.

## Files Likely to Need Attention During Sync
- `README.md`, `AGENTS.md` (project-specific docs)
- `internal/application.go` (module registration)
- `config/scheme.go`, `config/init.go` (configuration schema and defaults)
- `Makefile` (custom targets)

## Conflict Handling Tips
- Keep your local customizations; incorporate upstream changes where valuable.
- For documentation, prefer keeping your project-specific content while cherry-picking improvements from the template.
- For code, apply upstream fixes and enhancements while preserving your logic. Test after resolving conflicts.

## Template Version Tracking
- `.template-version` tracks:
  - `template_repo`: template URL
  - `last_sync_version`: tag/branch/commit merged
  - `last_sync_date`: UTC date of merge
  - `last_sync_commit`: commit SHA merged
- The template repo itself ignores this file; downstream repos keep it to know their sync point.

## Troubleshooting
- **Working tree not clean**: Commit or stash changes before running sync commands.
- **Remote missing**: Re-run `make template-setup` to ensure `template` remote exists.
- **Conflicts during merge**: Resolve manually, then continue with testing and commit.
- **Specific file diff**: Use `git diff template/main -- path/to/file` (or replace `template/main` with a tag).

## Good Practices
- Sync regularly to reduce conflicts.
- Keep template syncs in dedicated commits/PRs.
- Run tests after syncing.
- Review diffs before merging to understand incoming changes.

## Reference: Script Commands
- `scripts/template-sync.sh setup`
- `scripts/template-sync.sh status`
- `scripts/template-sync.sh fetch`
- `scripts/template-sync.sh diff [ref]`
- `scripts/template-sync.sh sync [ref]`

These are invoked via the Makefile targets for convenience.
