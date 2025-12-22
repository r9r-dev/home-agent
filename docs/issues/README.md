# Architecture Improvement Issues

This directory contains detailed issue specifications for improving the Home Agent codebase.
Each file corresponds to a GitHub issue that can be created.

## Issue Index

### P1 - High Priority

| # | Title | Component | Effort |
|---|-------|-----------|--------|
| 001 | [Split database.go into repositories](001-split-database-repositories.md) | Backend | Medium |
| 002 | [Implement database migrations](002-database-migrations.md) | Backend | Medium |
| 003 | [Extract prompt builders from chat.go](003-extract-prompt-builders.md) | Backend | Medium |
| 004 | [Create shared backend types](004-shared-backend-types.md) | Backend | Low |
| 005 | [Create shared frontend types](005-shared-frontend-types.md) | Frontend | Low |

### P2 - Medium Priority

| # | Title | Component | Effort |
|---|-------|-----------|--------|
| 006 | [Add repository interfaces](006-repository-interfaces.md) | Backend | Medium |
| 007 | [Split ChatWindow.svelte](007-split-chatwindow.md) | Frontend | Medium |
| 008 | [Add domain error types](008-domain-errors.md) | Backend | Low |
| 009 | [Extract backend constants](009-extract-backend-constants.md) | Backend | Low |
| 013 | [Refactor processMessage](013-refactor-process-message.md) | Proxy SDK | Medium |
| 016 | [Add unit tests](016-unit-tests.md) | Backend | High |
| 017 | [Create Svelte custom hooks](017-svelte-custom-hooks.md) | Frontend | Medium |
| 019 | [Extract frontend constants](019-frontend-constants.md) | Frontend | Low |

### P3 - Low Priority

| # | Title | Component | Effort |
|---|-------|-----------|--------|
| 010 | [Extract configuration](010-configuration-validation.md) | Backend | Low |
| 011 | [Add OpenAPI specification](011-openapi-specification.md) | Backend | High |
| 012 | [Clean up main.go](012-cleanup-main-go.md) | Backend | Medium |
| 014 | [Extract ExecutionContext class](014-execution-context-class.md) | Proxy SDK | Low |
| 015 | [Standardize logging](015-structured-logging.md) | Backend | Medium |
| 018 | [Split chatStore.ts](018-split-chat-store.md) | Frontend | Medium |

## Good First Issues

These issues are suitable for new contributors:

- [009 - Extract backend constants](009-extract-backend-constants.md)
- [019 - Extract frontend constants](019-frontend-constants.md)

## Dependencies

Some issues depend on others:

```
001 (Split database.go)
 └── 002 (Database migrations)
 └── 006 (Repository interfaces)
      └── 016 (Unit tests)

003 (Extract prompt builders)
 └── 004 (Shared backend types)

007 (Split ChatWindow)
 └── 017 (Custom hooks)
 └── 005 (Shared frontend types)

010 (Configuration)
 └── 012 (Clean up main.go)
```

## Creating Issues

To create these as GitHub issues, use the `gh` CLI:

```bash
# Create a single issue
gh issue create \
  --title "Split database.go into domain repositories" \
  --label "priority: P1" \
  --label "type: refactoring" \
  --label "component: backend" \
  --body-file docs/issues/001-split-database-repositories.md

# Or create all issues with a script
for file in docs/issues/[0-9]*.md; do
  # Extract title from first line
  title=$(head -1 "$file" | sed 's/# //')
  # Extract labels from file
  # ... create issue
done
```

## References

- [ARCHITECTURE_REVIEW.md](../../ARCHITECTURE_REVIEW.md) - Full architecture analysis
- [CLAUDE.md](../../CLAUDE.md) - Project documentation
