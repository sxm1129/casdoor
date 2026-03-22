# Casdoor 演进路线图 -- 三维度分析

> 日期: 2026-03-22
> 视角: 技术专家 + 产品专家
> 基线: 已完成 Route A 重构 (RBAC JOIN + UserIdentity 迁移) + 10 轮审计修复

---

## 当前已完成的基础工程

| 已修复项                | Commit          |
|------------------------|-----------------|
| RBAC LIKE -> JOIN 重写  | `79dff437`      |
| User 84 列 IdP -> UserIdentity 表 | `79dff437` |
| Role/Permission CRUD 映射表同步  | `adca9004` |
| CheckUserPermission bug     | `07bb698a`  |
| MasterPassword 安全加固       | `fd39ca8b`  |
| SQL Injection 字段校验        | `fd39ca8b`  |
| DB 连接池配置                | `fd39ca8b`  |
| JWT 敏感字段清除              | `ecb7ecc4`  |
| CORS + Auth return 修复      | `64274dad`  |
| fmt.Println -> logs.Error    | `6b4c2388`  |

以下分析均基于**当前状态**（而非原始开源版本），不再重复已完成的项。

---

## 维度一: 新增功能迭代

### 1.1 [高价值] Webhook 可靠投递引擎

**现状**: `SendWebhooks` 已异步 goroutine 化，但无重试、无签名、无死信队列。

**建议实现**:

| 特性 | 方案 |
|------|------|
| 重试 | 指数退避 3 次 (1s/4s/16s)，失败记录到 `webhook_delivery` 表 |
| 签名 | 对 payload 做 HMAC-SHA256，Header `X-Casdoor-Signature` |
| 幂等 | 每次投递生成 UUID delivery_id，接收方去重 |
| 死信 | 3 次失败后标记 `failed`，管理后台可手动重发 |
| 超时 | 可配置 per-webhook (默认 10s) |

**工作量**: ~3-5 天  
**业务价值**: 企业集成场景（用户创建推送到 HR 系统、钉钉群通知等）的基础设施。

---

### 1.2 [高价值] 结构化审计日志

**现状**: `Record` 表是 API 级日志，无法回答 "谁改了什么字段"。

**建议实现**:
- 新增 `audit_log` 表: `who, when, target_type, target_id, action, field_changes (JSON diff)`
- 在 `UpdateUser` / `UpdateRole` / `UpdatePermission` 等核心函数中，对比 old/new 对象生成 diff
- 管理后台新增 "操作记录" 页面，支持按用户/时间/操作类型筛选
- MasterPassword 使用时打上特殊标记 `loginMethod: "master_password"`

**工作量**: ~1 周  
**业务价值**: 等保 2.0/3.0 合规硬性要求；内部安全事件溯源。

---

### 1.3 [中价值] 用户自助服务门户

**现状**: 用户个人资料编辑在管理后台内，缺少独立的 C 端自助页面。

**建议实现**:
- 独立的 `/portal` 路由（无需管理后台权限）
- 个人信息编辑 + 密码重置 + MFA 绑定/解绑
- 绑定的社交账号管理（查看已绑定 IdP、解绑、新增绑定）
- 登录历史和活跃 Session（可踢除其他设备）

**工作量**: ~1-2 周  
**业务价值**: C 端用户自助管理，减轻管理员负担。

---

### 1.4 [中价值] API 密钥 / Service Account

**现状**: 仅支持 OAuth2 Client Credentials，无持久 API Key 机制。

**建议实现**:
- `api_key` 表: `owner, name, key_hash, scopes, expires_at, last_used_at`
- 管理后台可为 Application 创建多个 API Key
- 支持 `Authorization: Bearer casdoor_ak_xxxx` 认证方式
- Key 轮换：创建新 Key -> 老 Key 设定过期时间 -> 平滑迁移

**工作量**: ~3-5 天  
**业务价值**: 后端服务间调用、CI/CD 集成、无 OAuth 流程的 M2M (Machine-to-Machine) 场景。

---

### 1.5 [中价值] 用户导入/导出

**现状**: 有 Syncer 但无手动批量导入功能。

**建议实现**:
- 管理后台 "用户管理" 页增加 CSV/Excel 导入按钮
- 支持字段映射（上传后预览 + 选择对应列）
- 导出: 按 Organization 导出完整用户列表 (CSV/JSON)
- GDPR 数据可移植性条款要求

**工作量**: ~3-5 天

---

### 1.6 [低价值/长期] 事件总线抽象

**现状**: 仅 Webhook HTTP POST。

**未来方向**:
- 抽象 `EventPublisher` 接口
- 实现 Webhook / Kafka / RabbitMQ / Redis Pub/Sub 多后端
- 配置化: 不同事件类型路由到不同 Publisher

**工作量**: ~2-3 周  
**适合在用户量增长到需要 Event-Driven 时做**。

---

## 维度二: 现有功能提升

### 2.1 [P0] Schema 迁移机制引入

**现状**: `xorm.Sync2()` 无版本号、无回滚、不能删列。

**建议方案**:
- 引入 [goose](https://github.com/pressly/goose) 或 [golang-migrate](https://github.com/golang-migrate/migrate)
- `migrations/` 目录存放版本化迁移脚本 (SQL)
- 启动时检查 `schema_version` 表，自动执行待执行的迁移
- 保留 `Sync2` 仅用于开发模式 (`runmode=dev`)

**工作量**: ~1 周  
**风险**: **中** — 需要将现有 Sync2 生成的 Schema 作为 v1 基线

---

### 2.2 [P1] Controller 层测试补全

**现状**: controllers/ 目录 63 文件 **零测试**。

**分阶段策略**:

| 阶段 | 范围 | 优先级 |
|------|------|--------|
| Phase 1 | Auth 核心路径: `/api/login`, `/api/get-account`, `/api/logout` | 最高 |
| Phase 2 | Token 签发: `/api/login/oauth/access_token`, OIDC Endpoints | 高 |
| Phase 3 | CRUD 接口: User/Role/Permission 增删改查 | 中 |
| Phase 4 | Edge Cases: MFA flow, SAML callback, CAS protocol | 低 |

**使用 `httptest` + 真实 SQLite 内存库**，无需 Docker。

**工作量**: Phase 1-2 约 1-2 周；完整覆盖约 1 个月。

---

### 2.3 [P1] RBAC 查询缓存

**现状**: Route A 重构后 LIKE 已改为 JOIN，但每次权限校验仍需多次 DB 查询。

**建议方案**:
- 用户 Session 创建时，一次性预加载 `roles` + `permissions` 到 Session/Cache
- Role/Permission CRUD 变更时，广播 invalidation (Redis Pub/Sub 或进程内通知)
- 对于高频调用路径 (`CheckLoginPermission`)，引入进程内 LRU 缓存 (TTL 60s)

**工作量**: ~1 周  
**效果**: 权限校验从 3-5 次 DB 查询降为 0-1 次。

---

### 2.4 [P2] Beego -> 标准库 / Gin 渐进迁移

**现状**: Beego 2.x 生态活跃度低，社区萎缩。

**渐进策略** (不建议一次性重写):
- Phase 1: 新增路由使用 `net/http` Handler + adapter 注入到 Beego
- Phase 2: 逐步将高频路由从 Beego Controller -> 标准 `http.Handler`
- Phase 3: Session/Filter 去 Beego 化

**工作量**: 长期渐进，非一次性投入  
**紧迫度**: 低 — Beego 仍能工作，但长期技术债

---

### 2.5 [P2] 密码策略增强

**现状**: 支持密码长度校验，但缺少:
- 密码历史 (不能和最近 N 次相同)
- 密码过期策略 (N 天后强制修改)
- 账号锁定增强 (按 IP + 账号组合锁定)
- 密码强度实时反馈 (前端)

**工作量**: ~3-5 天

---

### 2.6 [P3] i18n 完善

**现状**: 支持多语言但翻译覆盖不完整。

**改进**:
- 审计所有 `i18n.Translate` 调用，确保后端错误消息全部走 i18n
- 前端管理后台补全中文翻译
- 登录页面字段标签和错误提示全量中文化

---

## 维度三: 前端 UI 完善

### 3.1 [P0] 登录页面全面升级

**现状**: 已有 Stitch 生成的设计稿（阶段 1 完成），但 React 代码层落地未做。

**待完成**:
- `web/src/auth/LoginPage.js` (75KB) — 当前虽已有部分改动，但整体仍是 Ant Design 原生表单
- 提取 Stitch 设计的 Split-Screen 布局、渐变背景、品牌展示区
- 响应式断点: Desktop (> 1024px) 双栏, Tablet/Mobile 单栏
- 社交登录按钮网格化排列（当前是竖列堆叠）
- 登录流程微动画（输入框 focus 过渡、按钮 loading 状态、错误抖动）

**涉及文件**:
- `web/src/auth/LoginPage.js`
- `web/src/auth/SignupPage.js`
- `web/src/auth/ForgetPage.js`
- `web/src/auth/PromptPage.js`
- `web/src/auth/MfaSetupPage.js`
- 新增 `web/src/auth/auth.css` 独立样式

**工作量**: ~1 周

---

### 3.2 [P0] Dashboard 数据大盘

**现状**: `web/src/basic/Dashboard.js` (252KB 改版后) 已有 ECharts 图表框架和卡片。

**待完善**:
- 接入真实后端数据 (`/api/get-dashboard`) — 当前 Dashboard API 已有，但前端消费不完整
- 核心指标卡片: 总用户数、今日新增、今日登录、在线用户、应用数量
- 趋势图: 7 天/30 天用户增长趋势、登录频次分布
- 接入系统健康状态卡片: 各 Application 的最近一次 Token 签发时间
- 暗色/亮色模式切换

**工作量**: ~3-5 天

---

### 3.3 [P1] 管理后台布局现代化

**现状**: Ant Design Pro 风格侧边栏 + 顶栏，样式陈旧。

**改进方向**:
- 侧边栏: 折叠式 + 图标模式，宽度可调
- 顶栏: 搜索框 (全局跳转到任何页面/用户/应用) + 通知铃铛 + 用户头像
- 面包屑导航优化
- 全局 ConfigProvider Token 调优: 圆角、间距、字体
- 表格页统一: 筛选器区域收纳、批量操作工具栏、行内快捷操作

**涉及文件**:
- `web/src/ManagementPage.js` (33KB — 主布局)
- `web/src/App.js` (27KB — 路由 + ConfigProvider)
- `web/src/App.less`
- 各 `*ListPage.js` (统一表格样式)

**工作量**: ~1-2 周

---

### 3.4 [P1] 用户编辑页瘦身

**现状**: `UserEditPage.js` **66KB** — 单文件画布式堆叠所有字段。

**改进**:
- 拆分为 Tab 页: "基本信息" / "安全设置" / "角色权限" / "社交账号" / "登录历史"
- 每个 Tab 独立组件 (`UserBasicInfo.js`, `UserSecuritySettings.js`, ...)
- 社交账号 Tab: 展示 `UserIdentity` 表绑定状态，支持管理员手动解绑/绑定
- 角色权限 Tab: 可视化展示用户拥有的 Role + 通过 Role 继承的 Permission
- 只读模式和编辑模式切换（避免误改）

**工作量**: ~1 周

---

### 3.5 [P2] 前端构建工具迁移

**现状**: `craco` (已停止维护) + `react-scripts` (CRA，已被 React 官方弃用)

**迁移路径**: craco -> Vite + @vitejs/plugin-react

**步骤**:
1. `npm init vite@latest` 初始化配置
2. 迁移 `craco.config.js` 中的 Less 配置 -> `vite.config.js` + `vite-plugin-imp`
3. 调整 import 路径 (CRA 的 `src/` alias)
4. 验证 HMR 和构建产出

**工作量**: ~2-3 天  
**收益**: 开发态 HMR 从 ~3s -> <500ms; 构建速度提升 5-10x。

---

### 3.6 [P2] 移动端响应式

**现状**: 管理后台几乎无移动端适配。

**核心改进**:
- 侧边栏: 移动端变为底部 Tab Bar 或 Drawer
- 表格: 移动端切换为卡片列表视图
- 登录页已在 3.1 中覆盖响应式

**工作量**: ~1 周

---

### 3.7 [P3] 全局暗色模式

**现状**: 无暗色模式。

**方案**: Ant Design 5.x 原生支持 `algorithm: theme.darkAlgorithm`，只需 ConfigProvider 级别切换 + 自定义 CSS 变量补充。

**工作量**: ~2-3 天

---

## 优先级总览

| 阶段 | 项目 | 维度 | 工作量 | 价值 |
|------|------|------|--------|------|
| **W1** | Schema 迁移引入 | 功能提升 | 1 周 | 升级安全性 |
| **W1** | 登录页 UI 落地 | 前端 | 1 周 | 用户第一印象 |
| **W2** | Dashboard 数据真实接入 | 前端 | 3-5 天 | 管理员体感 |
| **W2** | Webhook 可靠投递 | 新功能 | 3-5 天 | 企业集成 |
| **W2** | Controller 测试 Phase 1 | 功能提升 | 1 周 | 代码正确性 |
| **W3** | 结构化审计日志 | 新功能 | 1 周 | 合规 |
| **W3** | UserEditPage 拆分 | 前端 | 1 周 | 可维护性 |
| **W3** | RBAC 查询缓存 | 功能提升 | 1 周 | 性能 |
| **W4** | 管理后台布局重构 | 前端 | 1-2 周 | 专业感 |
| **W4** | 用户自助门户 | 新功能 | 1-2 周 | C 端体验 |
| **W5** | API Key / Service Account | 新功能 | 3-5 天 | M2M 场景 |
| **W5** | Vite 构建迁移 | 前端 | 2-3 天 | 开发体验 |
| **长期** | Beego 渐进迁移 | 功能提升 | 持续 | 技术债 |
| **长期** | 事件总线抽象 | 新功能 | 2-3 周 | 架构演进 |
