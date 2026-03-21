# Casdoor 作为统一用户中心的可行性分析报告

> 结论：**完全可行，且 Casdoor 天然就是为此场景设计的。**

---

## 一、项目概述

Casdoor 是一个开源的 **AI-first Identity and Access Management (IAM)** 平台，由 Casbin 社区维护。
技术栈：Go (Beego) + React + XORM，支持 MySQL/PostgreSQL/SQLite/MSSQL。

**核心定位**：集中式身份认证与授权网关，天然支持多系统共享一套用户体系。

---

## 二、核心架构分析

### 2.1 多租户模型（Organization）

```
Organization (租户)
  ├── User (用户, 属于某个 Organization)
  ├── Application (应用/系统, 绑定某个 Organization)
  ├── Role (角色, Organization 级别)
  ├── Permission (权限, 基于 Casbin)
  └── Group (用户组)
```

| 概念 | 作用 | 统一用户中心映射 |
|------|------|-----------------|
| Organization | 租户隔离单元 | 你的公司/团队 |
| User | 用户实体 | 统一用户 |
| Application | OAuth2/OIDC 客户端 | 每个接入系统 |
| Role | RBAC 角色 | 跨系统角色 |
| Permission | 细粒度权限 | 资源访问控制 |

**关键优势**：用户属于 Organization，Application 也属于 Organization。同一个 Organization 下的所有 Application 天然共享同一套用户。

### 2.2 用户模型（User）

User 结构体包含 **247 个字段**，覆盖：

- **基础信息**：用户名、邮箱、手机、地址、身份证等
- **社交账号绑定**：60+ 第三方 IdP（GitHub, Google, WeChat, DingTalk, Lark, Azure AD 等）
- **安全**：密码加密类型、MFA (TOTP/SMS/Email/Push/Radius)、WebAuthn
- **RBAC**：内嵌 Roles/Permissions/Groups
- **自定义字段**：10 个 Custom 字段 + Properties Map
- **审计**：登录时间、登录 IP、注册来源

### 2.3 认证协议栈

| 协议 | 状态 | 说明 |
|------|------|------|
| OAuth 2.0 / 2.1 | 完整实现 | authorization_code, password, client_credentials, implicit, device_code, token-exchange, refresh_token |
| OIDC | 完整实现 | Discovery, JWKS, UserInfo, End Session |
| SAML 2.0 | IdP 模式 | 支持 SP-initiated 和 IdP-initiated SSO |
| CAS | 完整实现 | v1/v2/v3 协议 |
| LDAP | 服务端 | 内建 LDAP Server，支持 LDAP 认证 |
| RADIUS | 服务端 | 内建 RADIUS Server |
| SCIM | 服务端 | 标准用户同步协议 |
| WebAuthn | 完整实现 | FIDO2 无密码登录 |
| Kerberos | 支持 | 企业级单点登录 |

### 2.4 权限模型（基于 Casbin）

Permission 模型支持：
- **RBAC**：User -> Role -> Permission 层级关系
- **ABAC**：基于属性的访问控制
- **资源级控制**：精确到 Application 级别的资源授权
- **Domain 隔离**：多域权限隔离
- **Model 自定义**：支持自定义 Casbin Model

---

## 三、接入方案分析

### 3.1 标准接入方式

```
┌───────────┐  OAuth2/OIDC  ┌──────────┐
│  系统 A   │ ────────────> │          │
│ (Web App) │               │          │
└───────────┘               │          │
                            │ Casdoor  │  <──>  Database
┌───────────┐  SAML/CAS     │ (统一    │
│  系统 B   │ ────────────> │  用户    │
│ (内部系统) │               │  中心)   │
└───────────┘               │          │
                            │          │
┌───────────┐  LDAP/RADIUS  │          │
│  系统 C   │ ────────────> │          │
│ (Legacy)  │               │          │
└───────────┘               └──────────┘
```

**每个接入系统**只需在 Casdoor 中创建一个 Application，获得 `ClientId` + `ClientSecret`，然后通过标准协议对接。

### 3.2 官方 SDK 生态

Casdoor 提供 20+ 语言/框架的官方 SDK：
- **后端**：Go, Java, Node.js, Python, PHP, .NET, Rust, Ruby, Dart, Elixir
- **前端**：React, Vue, Angular, Flutter, uni-app
- **移动端**：iOS, Android, React Native
- **框架集成**：Spring Boot, Django, NextAuth.js, Nuxt

### 3.3 用户同步（Syncer）

Casdoor 内建强大的用户同步功能，支持：
- Database Syncer（直接从现有系统数据库同步用户）
- Active Directory / Azure AD / Google Workspace
- DingTalk / Lark / WeCom（企业微信）
- Okta / Keycloak / SCIM
- 定时自动同步（Cron）

这意味着可以**零侵入地将现有系统的用户迁移到统一用户中心**。

---

## 四、实施路径建议

### Phase 1: 部署 + 初始配置
1. Docker 部署 Casdoor（`docker-compose.yml` 已就绪）
2. 创建 Organization（代表你的公司/团队）
3. 配置密码策略、MFA 策略、登录界面主题

### Phase 2: 存量用户迁移
1. 使用 Syncer 从各现有系统同步用户到 Casdoor
2. 映射用户字段（邮箱/手机/用户名去重合并）
3. 验证用户数据完整性

### Phase 3: 系统接入
1. 为每个系统创建 Application（获取 ClientId/ClientSecret）
2. 选择合适的协议（推荐 OAuth2/OIDC 用于 Web，LDAP 用于 Legacy）
3. 使用官方 SDK 集成到各系统
4. 配置 SSO 单点登录

### Phase 4: 权限统一管理
1. 定义全局 Role（如 admin, manager, viewer）
2. 配置 Permission 绑定 Application 资源
3. 实现跨系统权限联动

---

## 五、风险评估与注意事项

| 风险 | 等级 | 缓解措施 |
|------|------|----------|
| 单点故障 | 中 | 高可用部署（多实例 + Redis Session + 负载均衡） |
| 数据迁移冲突 | 中 | 先在测试环境验证用户合并逻辑 |
| 登录 UI 定制 | 低 | Casdoor 支持每个 Application 独立定制登录页面主题、CSS、Logo |
| 现有系统改造量 | 低-中 | SDK 封装良好，标准 OAuth2 流程，改造量可控 |
| 性能瓶颈 | 低 | Go 原生高性能 + Redis Session 缓存 |
| 密码兼容 | 低 | 支持 12+ 种密码哈希算法（pbkdf2, bcrypt, sha256-salt 等） |

---

## 六、结论

> [!IMPORTANT]
> Casdoor **完全可行**作为统一用户中心。它不仅仅是"可以做"，而是**天然就是为这个场景设计**的开源 IAM 平台。

**核心优势总结**：
1. **协议覆盖全面** - OAuth2/OIDC/SAML/CAS/LDAP/RADIUS/SCIM，几乎能对接任何系统
2. **多租户原生支持** - Organization 隔离，同组织用户天然共享
3. **SDK 生态丰富** - 20+ 语言 SDK，接入成本低
4. **用户同步内建** - Syncer 支持从 DB/AD/钉钉/飞书等迁移存量用户
5. **权限引擎成熟** - 基于 Casbin 的 RBAC/ABAC，支持跨系统权限管理
6. **社交登录完备** - 60+ 第三方登录，微信/钉钉/飞书等国内生态全覆盖
7. **安全能力强** - MFA/WebAuthn/密码策略/IP 白名单/审计日志
8. **开源 + 活跃社区** - Apache 2.0 协议，GitHub 12k+ Stars

**推荐方案**：直接在此项目基础上二次开发（如需），或直接部署使用并通过标准协议接入各系统。
