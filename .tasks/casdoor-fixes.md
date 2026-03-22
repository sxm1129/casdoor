# Task: casdoor-fixes

## 目标
修复 Casdoor 统一用户中心架构与安全级别的核心缺陷 (P0-P2)。

## Checklists
- [x] Phase 0: 核心漏洞与稳定性修复 (P0)
  - [x] 1. `object/organization.go` & `object/check.go` 禁用 MasterPassword 后门
  - [x] 2. `object/check.go:473` 修复 `CheckUserPermission` 的返回 error bug
  - [x] 3. `object/check.go:575` 修复 `CheckApiPermission` 的空匹配策略，默认拒绝改为一致的 allow/deny behavior
  - [x] 4. `object/record.go:163` 将 `AddRecord` 的 `panic` 替换为 error logs
  - [x] 5. `object/ormer.go` & `conf/app.conf` 增加数据库连接池配置
- [x] Phase 1: 纵深防御与异步化 (P1/P2)
  - [x] 6. `object/group.go` SQL Injection 防止，白名单验证 `field`
  - [x] 7. `object/record.go` Webhook 的 `SendWebhooks` 放入 Goroutine 异步化
  - [x] 8. `object/session.go` 修改 100 限制，开放到 1000 并在配置中可选
