# Production Audit — 2026-03-22 Phase 2 (Fixes)

## Scope
All `.go` files modified for security and stability in `task/casdoor-fixes` (commit `fd39ca8b`).

## File Inventory (Priority)

### P0 — Auth, Data Mutation
| File | Lines | Focus |
|------|-------|-------|
| check.go | 813 | MasterPassword, UserPermission, ApiPermission |
| record.go | 368 | Webhooks, logging panics |

### P1 — Business Logic
| File | Lines | Focus |
|------|-------|-------|
| form.go / session.go | 266 | Session limiting |
| group.go | 469 | SQL Injection prevention |
| ormer.go | 518 | DB limit |
| organization.go | 634 | MasterPassword disabling |

**Total: 6 files, ~3,000 lines**

## Rounds
- [x] Round 1

## Findings (Round 1)

### Bugs
1. **BUG-01 (group.go:327)** `GetGroupUserCount`: ~~SQL tableNamePrefix 拼接错误~~ — 已在 Route A 重构 (`79dff437`) 中使用 `prefixedUserTable` 变量自行修正，当前代码无此问题。

### Smells
1. **SMELL-01 (record.go:159,166)** `AddRecord`: webhook goroutine 和 DB insert 失败处替换为 `logs.Error`，修复于 commit `fix(audit-smell01)`。
