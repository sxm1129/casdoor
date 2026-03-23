# Casdoor 改进分析报告

> 基于代码库深度扫描: 87 前端页面, 69 控制器, 151 对象文件

---

## 一、新增功能迭代

### 1. Syncer Group Sync — 完成 20 个 TODO 桩函数
**优先级**: P0 | **工作量**: 中 | **价值**: 高

代码库中有 **20 个 TODO** 分布在 10 个 Syncer 实现中:

| Syncer | 缺失功能 |
|--------|----------|
| Keycloak | `SyncGroups`, `SyncGroupMembers` |
| Lark (飞书) | `SyncGroups`, `SyncGroupMembers` |
| WeCom (企微) | `SyncGroups`, `SyncGroupMembers` |
| Active Directory | `SyncGroups`, `SyncGroupMembers` |
| Okta | `SyncGroups`, `SyncGroupMembers` |
| Azure AD | `SyncGroups`, `SyncGroupMembers` |
| SCIM | `SyncGroups`, `SyncGroupMembers` |
| Database | `SyncGroups`, `SyncGroupMembers` |
| AWS IAM | UserId→UserName mapping for group sync |

> [!IMPORTANT]
> 这是生产级 IAM 系统的核心能力缺失。企业用户需要完整的组同步才能实现 SSO + RBAC 闭环。

---

### 2. 审计日志归档 & 保留策略
**优先级**: P1 | **工作量**: 小 | **价值**: 高

当前 `record.go` 只有写入，无清理机制。`Record` 表会无限增长。

**建议**:
- 新增 `CleanupOldRecords(retentionDays int)` 定时任务
- Organization 级别配置 `AuditRetentionDays`
- 可选: 归档到 S3/OSS 后再删除

---

### 3. API Key 生命周期管理
**优先级**: P1 | **工作量**: 小 | **价值**: 中

当前 `api_key.go` 仅实现了基础 CRUD。缺少:
- API Key 到期时间 + 自动轮换
- 按 Key 的 API 调用计数统计
- Key 的权限范围限定 (scope)

---

### 4. Webhook 重试 & 死信队列
**优先级**: P2 | **工作量**: 中 | **价值**: 高

当前 `webhook_worker.go` 仅同步触发一次, 失败后只记录日志:
- 增加指数退避重试 (3 次)
- 失败超过阈值后进入死信队列
- 管理界面展示失败投递并支持手动重发

---

### 5. 用户自助服务 API
**优先级**: P2 | **工作量**: 小 | **价值**: 中

- `POST /api/request-account-deletion` — 用户申请删除账户 (GDPR 合规)
- `GET /api/export-my-data` — 用户导出个人数据 (数据可携权)
- `POST /api/revoke-all-sessions` — 用户一键注销所有会话

---

### 6. 登录事件分析 & 异常告警
**优先级**: P2 | **工作量**: 中 | **价值**: 高

基于 `record.go` 已有的审计日志:
- 检测异常登录模式 (异地登录、暴力破解)
- 自动触发 MFA 升级或临时锁定
- Dashboard 展示登录趋势图 (复用已有 echarts)

---

### 7. Organization 配额管理
**优先级**: P3 | **工作量**: 小 | **价值**: 中

`util.go` 已有 `checkQuotaForUser/Application/Organization/Provider`, 但均读配置文件。改为:
- 每个 Organization 可单独配置配额
- 超配额时前端实时提示
- 管理面板展示配额使用率

---

### 8. 数据库迁移框架增强
**优先级**: P3 | **工作量**: 中

当前 `migrate.go` + `migration.go` 仅处理 legacy `user_role` 迁移。建议:
- 引入版本化迁移 (类似 `golang-migrate`)
- `migrations/` 目录 + SQL 文件模式
- 自动记录已执行的迁移版本

---

## 二、现有功能提升

### 1. 后端巨型文件拆分
**优先级**: P0 | **工作量**: 中

| 文件 | 大小 | 建议 |
|------|------|------|
| `object/user.go` | 43KB / 1580 行 | 拆分为 `user_crud.go`, `user_query.go`, `user_filter.go` |
| `object/token_oauth.go` | 42KB | 拆分 OAuth 各 grant type 为独立文件 |
| `controllers/auth.go` | 47KB / 1515 行 | 按登录方式拆分: `auth_password.go`, `auth_oauth.go`, `auth_saml.go` |
| `object/user_util.go` | 29KB | 提取辅助函数到更小的文件 |

---

### 2. 缓存层统一 & 监控
**优先级**: P1 | **工作量**: 小

当前存在 3 套独立缓存:
- `rbac_cache.go` — sync.Map + TTL
- `rule_cache.go` — 全局 map
- `site_cache.go` — 同步加载

**建议**: 统一为 `cache.go` 封装层, 增加:
- 缓存命中率 Prometheus 指标
- 可配置 TTL
- `/api/cache-stats` 管理端点

---

### 3. IsOrgAdminOfOwner 在关键 API 中落地
**优先级**: P1 | **工作量**: 小

已创建 `IsOrgAdminOfOwner()`, 但尚未应用到以下 API:
- `UpdateUser` / `DeleteUser` — 阻止跨组织操作
- `GetUsers` — org-admin 只能看到本组织用户
- RBAC 相关: `AddRole`, `DeleteRole`, `AddPermission`

---

### 4. Token 清理任务优化
**优先级**: P1 | **工作量**: 小

`token_cleanup.go` 已有定时清理, 但:
- 缺少指标 (清理了多少过期 token)
- 大表场景下 `DELETE` 性能问题 (应分批删除)
- 应增加 `LIMIT` + 循环删除策略

---

### 5. 密码策略 UI 配置化
**优先级**: P2 | **工作量**: 小

R7 已实现了密码历史 + 过期。但 Organization 编辑页尚未暴露:
- `PasswordHistoryLimit` (当前硬编码 5)
- `PasswordExpireDays` (已有字段, 需 UI 输入框)
- `PasswordComplexity` 规则配置 (已有 `check_password_complexity.go`, 需可视化)

---

### 6. LDAP/AD InsecureSkipVerify 可配置化
**优先级**: P2 | **工作量**: 极小

`syncer_activedirectory.go:151` 硬编码 `InsecureSkipVerify: true` + TODO 注释。应:
- 在 Syncer 配置中增加 `TlsSkipVerify` 布尔字段
- 前端 SyncerEditPage 添加开关控件

---

### 7. Prometheus 指标扩展
**优先级**: P2 | **工作量**: 小

当前 `prometheus.go` 仅有 `ApiLatency` 和 `ApiCounter`。建议增加:
- `LoginAttempts` (成功/失败, 按 org 分组)
- `TokenIssued` / `TokenRevoked`
- `WebhookDeliveryStatus` (成功/失败/重试)
- `CacheHitRate` (RBAC / Rule / Site)

---

### 8. 控制器测试覆盖扩展
**优先级**: P3 | **工作量**: 中

R1 已建立基础设施 (8 个测试)。下一步:
- 添加 OAuth 授权码流程端到端测试
- SAML SP/IdP 交互测试
- Token 刷新 + 撤销流程测试
- Permission enforcement 测试

---

## 三、前端 UI 完善

### 1. Class Component → Hooks 迁移
**优先级**: P1 | **工作量**: 大 (渐进式)

当前有 **50+ 个 Class Component** (使用 `extends React.Component`)。现代 React 生态已全面转向 Hooks。

**建议渐进路径**:
1. 新组件一律使用 Hooks
2. 高频修改的组件优先迁移: `UserEditPage`, `ApplicationEditPage`, `OrganizationEditPage`
3. Table 组件批量迁移 (模式统一)

---

### 2. Setting.js 巨型文件拆分
**优先级**: P0 | **工作量**: 中

`Setting.js` 有 **80KB / ~2300 行**, 是全项目最大的单文件。包含:
- 主题配置、语言、URL 工具函数、全局状态

**建议拆分**:

| 新文件 | 内容 |
|--------|------|
| `utils/theme.js` | 主题、暗色模式、颜色 |
| `utils/url.js` | URL 构建、路由辅助 |
| `utils/format.js` | 日期、数字格式化 |
| `utils/auth.js` | 权限判断、isAdmin 等 |

---

### 3. ApplicationEditPage 拆分
**优先级**: P1 | **工作量**: 中

`ApplicationEditPage.js` 有 **83KB**, 是前端第二大文件。建议:
- 按功能区域拆分为 Tab + 子组件 (参考已完成的 UserEditPage)
- OAuth 配置、UI 定制、Provider 绑定分别独立

---

### 4. Dashboard 数据可视化增强
**优先级**: P1 | **工作量**: 中

当前 Dashboard (`/api/get-dashboard`) 只返回 30 天统计数。建议:
- 用 echarts 绘制: 用户增长趋势、登录频率热力图、活跃 Provider 排名
- 可选时间范围 (7天/30天/90天)
- 组织级 vs 全局视图切换

---

### 5. 暗色模式支持
**优先级**: P2 | **工作量**: 中

`index.html` 已引入 Tailwind 并配置了 `darkMode: "class"`, 但:
- Ant Design 组件未配置暗色主题 token
- App.less 无暗色变量
- 用户偏好未持久化

**建议**: 使用 Ant Design 5 的 `ConfigProvider` + `theme.algorithm` 切换

---

### 6. 表格组件统一化
**优先级**: P2 | **工作量**: 中

List 页面 (UserListPage, RoleListPage 等) 存在大量重复代码。建议:
- 抽取 `GenericListPage` 高阶组件
- 统一搜索、分页、排序逻辑
- 列配置声明化 (类似 ProTable)

---

### 7. 移动端适配
**优先级**: P2 | **工作量**: 中

`Setting.isMobile()` 散落在各处做条件渲染, 但:
- 侧边栏在移动端无折叠
- 表格在小屏溢出
- 登录页布局未针对移动端优化

---

### 8. 前端国际化完善
**优先级**: P3 | **工作量**: 小

`locales/` 目录存在但部分 key 缺失翻译 (fallback 到英文)。建议:
- CI 中增加 i18n key 覆盖率检查
- 补全中文翻译
- 考虑 RTL 布局支持 (阿拉伯语等)

---

## 建议执行顺序

| 阶段 | 目标 | 预估周期 |
|------|------|----------|
| Sprint 1 | 审计日志归档 + IsOrgAdmin落地 + 密码策略UI + Token清理优化 | 2-3 天 |
| Sprint 2 | Setting.js 拆分 + Dashboard 数据可视化 | 3-5 天 |
| Sprint 3 | Syncer Group Sync (飞书/企微优先) | 5-7 天 |
| Sprint 4 | ApplicationEditPage 拆分 + 暗色模式 | 3-5 天 |
| Sprint 5 | Webhook 重试 + Prometheus 扩展 + API Key 增强 | 3-5 天 |
