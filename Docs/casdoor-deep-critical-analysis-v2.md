# Casdoor 统一用户中心 -- 深度批判性分析 (v2)

> **目的**: 本文档是对 `casdoor-critical-analysis.md` 的深化版本，基于逐行代码审计，从**架构设计、安全攻击面、工程质量、运维可观测性**四个维度进行 Red Team 级别的批判性分析。所有结论均附精确代码行号。
>
> **基线**: 分析基于本地 fork (2026-03-21 快照)。

---

## 一、安全攻击面（Security Attack Surface）

### 1.1 [严重] MasterPassword -- 组织级万能密码后门

`Organization` 模型包含一个 `MasterPassword` 字段 (`organization.go:75`)，在 `CheckPassword()` 中优先于用户真实密码进行校验：

```go
// object/check.go:263-267
if organization.MasterPassword != "" {
    if password == organization.MasterPassword ||
       credManager.IsPasswordCorrect(password, organization.MasterPassword, organization.PasswordSalt) {
        return resetUserSigninErrorTimes(user)
    }
}
```

**风险评估**:
- **任何知道 MasterPassword 的人可以登录该 Organization 下的任意用户账号**
- MasterPassword 在 API 返回时被 mask 为 `***` (`organization.go:193-194`)，但在数据库中以哈希形式存储 -- 如果攻击者获得 DB 读权限 + 知道密码类型/salt，可以尝试离线破解
- 没有使用 MasterPassword 登录时的特殊审计标记（Record 中无法区分是正常登录还是 MasterPassword 登录）
- **无法在代码层面禁用此功能** -- 唯一的防御是不设置 MasterPassword

**严重等级**: **P0/Critical** -- 等保/GDPR 合规中，这是一个无法接受的设计。

### 1.2 [严重] CheckUserPermission Bug -- 永久返回 Error

```go
// object/check.go:473
return hasPermission, errors.New(i18n.Translate(lang, "auth:Unauthorized operation"))
```

**无论 `hasPermission` 是 `true` 还是 `false`，都返回一个 `error`**。

这意味着所有调用方必须忽略 error 返回值，只看 bool。这违反了 Go 惯例，也意味着：
- 如果某个新开发者正确地检查 `err != nil`，会导致已授权操作被拒绝
- 日志中会充斥大量虚假的 "Unauthorized operation" 错误

### 1.3 [高] CheckApiPermission 死逻辑 -- 默认全部拒绝

```go
// object/check.go:575-578
if allowPermissionCount > 0 && denyPermissionCount == 0 {
    return false, nil  // no-match = deny (合理)
}
return false, nil  // <-- 但这里：有 deny 规则存在时，no-match 也是 deny
```

**对比 `CheckLoginPermission`** (L686-689):
```go
if allowPermissionCount > 0 && denyPermissionCount == 0 {
    return false, nil
}
return true, nil  // <-- 这里是 allow
```

两个函数的默认行为**不一致**:
- `CheckApiPermission`: 当用户没命中任何规则时，**默认拒绝**（所有路径最终 return false）
- `CheckLoginPermission`: 当用户没命中任何规则时，**默认允许**（最终 return true）

这种不一致性会导致极难排查的权限问题。

### 1.4 [高] SQL 字段名注入 -- group.go

```go
// object/group.go:323
And(fmt.Sprintf("user.%s like ?", util.CamelToSnakeCase(field)), "%"+value+"%")

// object/group.go:349
And(fmt.Sprintf("%s.%s like ?", prefixedUserTable, util.CamelToSnakeCase(field)), "%"+value+"%")
```

`field` 参数经过 `CamelToSnakeCase` 转换后直接拼接到 SQL 语句中。虽然 `value` 使用参数化查询，但 **`field` 本身未经白名单校验**。

如果 API 层没有严格验证 `field` 参数值（目前 controller 层零测试覆盖），攻击者可以构造：
```
field = "1=1 OR name" → SQL: user.1=1_or_name like ?
```
实际 SQL 注入可能性取决于 `CamelToSnakeCase` 实现是否过滤特殊字符，但这属于**纵深防御缺失**。

### 1.5 [中] WebAuthn LIKE 查询用户枚举

```go
// object/user.go:510-518
if ormer.driverName == "postgres" {
    existed, err = ormer.Engine.Where(builder.Like{"\"webauthnCredentials\"", webauthId}).Get(&user)
} else {
    existed, err = ormer.Engine.Where("webauthnCredentials like ?", "%"+webauthId+"%").Get(&user)
}
```

WebAuthn Credential ID 使用 `LIKE '%...%'` 全表扫描匹配。如果 Credential ID 较短，可能会误匹配到其他用户的 Credentials（JSON blob 中的部分匹配）。

---

## 二、架构级缺陷（Architectural Flaws）

### 2.1 [严重] RBAC 查询的 N+1 瀑布效应

一次权限校验的调用链（以 `CheckLoginPermission` 为例）：

```
CheckLoginPermission(userId, app)
  → GetPermissions(org)                    // 查全部 permissions (SELECT * FROM permission)
  → for each permission:
      → permission.isUserHit(userId)       // 内存遍历
      → permission.isRoleHit(userId)       // ← 关键瓶颈
          → getRolesByUser(userId)          // 每 permission 调一次！
              → getRolesByUserInternal(userId)
                  → GetUser(userId)         // DB: SELECT * 247 cols
                  → ormer.Engine.Where("r.users like ?", ...)  // LIKE 全表扫描
                  → for matched roles: util.InSlice() 二次验证
              → GetAncestorRoles(allRolesIds...)
                  → GetRoles(owner)        // DB: SELECT * FROM role WHERE owner=?
                  → for each: containsRole() 递归遍历
      → getPermissionEnforcer(permission)  // 初始化 Casbin Enforcer
      → enforcer.Enforce(...)
```

**症结**: `permission.isRoleHit(userId)` 在 `for` 循环中被**每个 permission 调用一次**，而其内部又做了完整的 DB 查询 + 角色树遍历。

**假设场景**: Organization 有 50 个 Permission、用户属于 3 个 Role、每个 Role 有 5 层继承关系：
- Permission 查询: 1 次
- `getRolesByUser` 被调用: ~50 次（每个 permission 中的 isRoleHit）
- 每次 `getRolesByUser` 内部: 至少 3 次 DB 查询
- **总计 DB 查询: ~150+ 次**

**对比**: Keycloak 的 RBAC 预加载机制在用户 Session 创建时一次性计算角色树并缓存。

### 2.2 [严重] 全局单例 ORM + 无连接池配置

```go
// object/ormer.go:44-45
var (
    ormer *Ormer = nil  // 全局唯一
)
```

```go
// object/ormer.go:251-270 — open() 方法
func (a *Ormer) open() error {
    engine, err := xorm.NewEngine(a.driverName, dataSourceName)
    // ... 没有任何连接池配置
    a.Engine = engine
    return nil
}
```

**缺失的配置**:
- `engine.SetMaxOpenConns()` -- 未调用
- `engine.SetMaxIdleConns()` -- 未调用
- `engine.SetConnMaxLifetime()` -- 未调用

XORM 默认值：`MaxOpenConns = 0`（无限制），`MaxIdleConns = 2`。这意味着在高并发场景下：
- 可能耗尽数据库连接
- 每次请求可能都创建新连接（因为 idle pool 才 2 个）

### 2.3 [严重] Sync2 Schema 迁移 -- 数据不一致风险

```go
// object/ormer.go:302-474 — createTable()
func (a *Ormer) createTable() {
    // 30+ 次 Sync2 调用，任何一次失败都 panic
    err := a.Engine.Sync2(new(Organization))
    if err != nil { panic(err) }
    // ...
}
```

**Sync2 的已知限制**:
1. **只能 ADD 列，不能 DROP 或 MODIFY** -- 如果你重命名了字段或改变了类型，旧列留在数据库中成为垃圾数据
2. **大表 ALTER TABLE 可能锁表** -- 247 列的 User 表在百万级数据时执行 Sync2 可能导致分钟级锁
3. **无版本号、无回滚** -- 无法知道当前 Schema 是哪个版本，也无法回退

### 2.4 [高] Webhook 同步阻塞请求路径

```go
// object/record.go:156-161
errWebhook := SendWebhooks(record)  // 同步调用
if errWebhook == nil {
    record.IsTriggered = true
} else {
    fmt.Println(errWebhook)  // 错误仅打印到 stdout
}
```

`SendWebhooks` 是**同步**调用，在请求的 AfterExec filter 中执行。如果 Webhook 目标 URL 响应慢或超时：
- **阻塞当前请求的响应**
- 无重试机制
- 无超时配置（依赖 HTTP 默认超时）
- 错误仅 `fmt.Println` 到 stdout，非结构化日志

更严重的是 `addRecord` 失败时直接 `panic`:
```go
// object/record.go:163-166
affected, err := addRecord(record)
if err != nil {
    panic(err)  // Record 写入失败 = 整个请求 panic
}
```

### 2.5 [中] User 宽表 -- 247 列 God Object

每个 IdP 占一个独立列 (`user.go:122-200` 约 80 个 IdP 字段)。问题：
- 新增 IdP 需要 ALTER TABLE ADD COLUMN
- 每次 `SELECT *` 加载 247 个字段（无投影查询优化）
- JSON 序列化传输冗余数据
- 不活跃的 IdP 列（如 `Oura`, `EveOnline`, `BattleNet`）永远为空但占用 Schema 空间

**对比正确设计**: 使用 `user_idp_link` 关联表 (`user_id, idp_type, idp_user_id`)。

---

## 三、代码质量 & 工程实践

### 3.1 [严重] Controller 零测试覆盖

| 目录 | 文件数 | 测试文件数 | 核心风险 |
|------|--------|-----------|---------|
| `controllers/` | 63 文件 | **0** | API 权限校验、参数验证零自动化测试 |
| `object/` | ~138 文件 | 10 | Token 流程、权限校验无集成测试 |
| `routers/` | ~10 文件 | **0** | Filter 链（CORS/Timeout/Auth）无测试 |
| `idp/` | 33 文件 | **0** | 第三方登录回调无模拟测试 |

**现实影响**: 任何 API 的权限校验逻辑变更都无法通过自动化手段验证正确性。这在 IAM 系统中是不可接受的。

### 3.2 [高] Panic 滥用

在生产代码路径中发现以下 panic 使用模式：

| 位置 | 触发条件 | 影响 |
|------|----------|------|
| `ormer.go:302-474` | DB Schema 同步失败 | 启动崩溃 x30 |
| `ormer.go:142-153` | Engine.Close() 失败 | GC 回收期间崩溃 |
| `record.go:165` | Record INSERT 失败 | 用户请求 panic |
| `user.go:49` | Enforcer 初始化失败 | 启动崩溃 |
| `role.go:209-211` | Batch INSERT Duplicate | 非重复错误 panic |

特别是 `record.go:165` -- 这意味着如果审计日志写入失败（如 DB 磁盘满），所有用户请求都会 panic。

### 3.3 [中] 错误处理反模式

大量函数的返回值模式不一致：

```go
// 模式一：返回空字符串表示无错误 (check.go)
func CheckUserSignup(...) string { return "" }  // "" = OK

// 模式二：标准 Go error (check.go)
func CheckPassword(...) error { return nil }

// 模式三：bool + error 但 error 始终非 nil (check.go:473)
func CheckUserPermission(...) (bool, error) {
    return hasPermission, errors.New("Unauthorized")  // 永远返回 error
}

// 模式四：bool 返回 + 内部 panic (record.go:142-168, role.go:203-213)
func AddRecord(record *Record) bool {
    // ... panic(err) 在内部
}
```

没有统一的错误处理策略。

### 3.4 [中] Session 管理硬编码限制

```go
// object/session.go:128-132
func removeExtraSessionIds(session *Session) {
    if len(session.SessionId) > 100 {
        session.SessionId = session.SessionId[(len(session.SessionId) - 100):]
    }
}
```

单个用户最多 100 个 Session ID，超出后截断（非 FIFO）。这个硬编码限制：
- 无法配置
- 截断时直接丢弃旧 Session，无通知
- 在 SSO 场景下（一个用户同时登录 20+ 系统），100 可能很快达到上限

---

## 四、产品 & 运维维度

### 4.1 审计日志不合规

`Record` 表仅记录 API 调用维度的日志，无法满足以下审计需求：
- "谁在什么时间修改了用户 X 的什么字段？" -- 不记录字段级变更
- "中间 MasterPassword 是否被使用过？" -- 无法区分
- 无 GDPR "数据访问日志" -- 无法回答 "谁查看了用户 X 的个人信息"
- `Record.Response` 字段被截断且格式非结构化 (`fmt.Sprintf`)

### 4.2 Webhook 无可靠性保证

| 特性 | Casdoor | 行业标准 |
|------|---------|---------|
| 重试 | 无 | 指数退避 3-5 次 |
| 幂等 | 无 (无 delivery ID) | 每次投递携带 UUID |
| 签名验证 | 无 | HMAC-SHA256 签名 |
| 死信队列 | 无 | DLQ + manual replay |
| 异步投递 | 否 (同步阻塞) | 异步 + 确认 |
| 超时 | 默认 HTTP 超时 | 可配置 per-webhook |

### 4.3 API 设计非标准

所有 API 使用动词式路径：
- `POST /api/login` (OK)
- `GET /api/get-users` (应为 `GET /api/users`)
- `POST /api/update-user` (应为 `PUT /api/users/:id`)
- `POST /api/delete-user` (应为 `DELETE /api/users/:id`)

这不是纯美学问题 -- 在 API Gateway 场景下，基于 HTTP Method 的路由规则无法正确匹配。

### 4.4 前端技术债

- 构建工具: craco (已停止维护，最后更新 2023 年)
- React 版本: 基于 CRA（react-scripts），已被 React 官方弃用
- 无服务端渲染，不支持 SEO（IAM 管理后台不需要，但登录页可能需要）

---

## 五、风险合并评估矩阵

| # | 问题 | 严重度 | 影响域 | 修复成本 | 独立修复? |
|---|------|--------|--------|----------|-----------|
| 1 | MasterPassword 万能密码 | **P0** | 安全/合规 | 低 (禁用/移除功能) | 是 |
| 2 | CheckUserPermission 返回 error bug | **P0** | 权限正确性 | 低 (1 行修复) | 是 |
| 3 | RBAC LIKE 查询 + N+1 瀑布 | **P1** | 性能/正确性 | 高 (需重构关联模型) | 否 |
| 4 | Controller 零测试 | **P1** | 代码正确性 | 高 (补充大量测试) | 是 |
| 5 | Sync2 无版本迁移 | **P1** | 升级安全性 | 中 (引入 goose/migrate) | 是 |
| 6 | User God Object 247 字段 | **P1** | 性能/可维护性 | 极高 (数据模型重构) | 否 |
| 7 | Webhook 同步阻塞 + 无重试 | **P2** | 可靠性/集成 | 中 (异步队列) | 是 |
| 8 | CheckApiPermission 死逻辑 | **P2** | 权限正确性 | 低 (修正返回值) | 是 |
| 9 | SQL 字段名注入 | **P2** | 安全 | 低 (加白名单) | 是 |
| 10 | Record panic | **P2** | 稳定性 | 低 (改 error return) | 是 |
| 11 | 无连接池配置 | **P2** | 性能 | 低 (3 行配置) | 是 |
| 12 | Session 硬编码 100 | **P3** | 边界情况 | 低 (改为配置) | 是 |
| 13 | API 非 RESTful | **P3** | 开发体验 | 极高 (全面重构) | 否 |
| 14 | 前端 craco 停止维护 | **P3** | 可维护性 | 中 (迁移 Vite) | 是 |

---

## 六、决策分析

### 6.1 "直接使用" 方案的真实代价

如果选择直接使用 Casdoor 作为统一用户中心，你需要**接受但清醒认识**以下事实：

**可以接受的**:
- API 非 RESTful -- 用 SDK 封装，不影响功能
- 前端 craco -- 管理后台 UI 不面向终端用户
- Session 100 上限 -- 大多数场景够用

**需要立即修复的**（即使"直接使用"也不可接受）:
1. **禁用 MasterPassword** -- 在 `app.conf` 或代码层面强制禁止设置
2. **修复 CheckUserPermission bug** -- 1 行代码
3. **修复 CheckApiPermission 默认行为** -- 确认意图后修改
4. **Record panic 改为 error** -- 防止审计日志故障导致服务雪崩

**需要在上生产前完成的**:
5. 配置 Redis Session（已支持，需配置）
6. 配置 DB 连接池参数
7. SQL 字段名白名单校验

### 6.2 "二次开发" 方案的工作量估算

| 修复项 | 预估工作量 | 风险 |
|--------|-----------|------|
| P0 修复 (1-2, 8, 10) | 1-2 天 | 低 |
| 连接池 + Session + SQL 白名单 | 1 天 | 低 |
| Webhook 异步化 | 3-5 天 | 中 |
| RBAC 性能优化 (关联表 + 缓存) | 2-3 周 | 高 (数据迁移) |
| DB 迁移机制引入 | 1 周 | 中 |
| User 模型拆分 | 3-4 周 | 极高 (全面重构) |
| Controller 测试补全 | 2-3 周 | 中 |
| **合计** | **2-3 人月** | |

### 6.3 "换方案" 的考量

| 方案 | 国内生态 | 工程质量 | 运维复杂度 | 推荐场景 |
|------|---------|---------|-----------|---------|
| **Casdoor** | 优秀 | 低 | 低 | 快速上线、国内社交登录、<50 万用户 |
| **Keycloak** | 差 | 高 | 高 (JVM) | 大规模企业、强合规、已有 Java 技术栈 |
| **Authentik** | 差 | 中 | 中 | 已有 Python 技术栈、需要灵活自定义 Flow |
| **Logto** | 一般 | 高 | 低 | 面向 C 端、TypeScript 全栈 |

---

## 七、最终结论与建议

### 核心判断

> Casdoor **功能覆盖面广、国内生态优秀**，但在**安全纵深防御、代码质量纪律、架构可扩展性**三个维度存在系统性缺陷。
>
> 对于一个 IAM 系统来说，"能用" 和 "安全可靠" 之间存在显著鸿沟。

### 推荐策略

**如果你的场景满足以下全部条件，推荐 Casdoor**:
1. 用户量 < 50 万
2. 接入系统 < 20 个
3. 对国内社交登录（微信/钉钉/飞书/支付宝）有硬性需求
4. 团队有 Go 背景、愿意投入 1-2 人月做 P0/P1 修复
5. 不需要通过等保三级或 GDPR 合规

**如果以下任一条件成立，建议换方案**:
1. 用户量预期超过百万
2. 有等保三级/GDPR 合规要求
3. 需要事件驱动架构（用户变更推送到下游系统）
4. 团队工程标准高，无法接受 "controller 零测试" 的代码质量
5. 需要企业级多租户数据物理隔离

### 如果继续使用 Casdoor 的最小修复清单

```
Phase 0 (Day 1-2):
  [x] 禁用 MasterPassword
  [x] 修复 CheckUserPermission return bug
  [x] Record panic → error
  [x] DB 连接池配置

Phase 1 (Week 1-2):
  [ ] SQL field 白名单
  [ ] Redis Session 配置
  [ ] Webhook 异步化（goroutine pool + 重试）

Phase 2 (Month 1-2):
  [ ] RBAC 关联表重构
  [ ] DB Migration 机制引入
  [ ] Controller 核心路径测试补全
```
