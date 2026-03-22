# Route A 彻底重构 -- 实施计划

> 解决 RBAC N+1 性能雪崩 + User 247 列神对象问题

---

## 背景

当前 `task/casdoor-refactor` 分支已完成模型定义和迁移代码，但核心查询重写和适配层未开始。本计划将完成剩余全部工作。

## User Review Required

> [!IMPORTANT]
> - `UserWithoutThirdIdp` struct 和 `getUserByWechatId` 中的 IdP 直接列引用将被改为查 `UserIdentity` 表
> - `UpdateUser` 函数中 776-790 行的 84 个 IdP 列名将从 columns 列表中移除
> - RBAC LIKE 查询改为 JOIN `UserRole`/`RolePermission`/`UserPermission` 映射表

---

## Proposed Changes

### Component 1: RBAC LIKE → JOIN 重写

---

#### [MODIFY] [role.go](file:///Users/hs/workspace/github/casdoor/object/role.go)

**`getRolesByUserInternal`** (L272-298):
- 当前: `WHERE r.users LIKE '%userId%'` + 内存二次验证
- 改为: `SELECT r.* FROM role r INNER JOIN user_role ur ON ur.role = CONCAT(r.owner,'/',r.name) WHERE ur.user = ?`
- 同时保留 groups LIKE 查询（groups 数量少，LIKE 可接受）

---

#### [MODIFY] [permission.go](file:///Users/hs/workspace/github/casdoor/object/permission.go)

3 个 LIKE 函数改为 JOIN:
- **`getPermissionsByUser`** (L329-344): JOIN `user_permission` 表
- **`GetPermissionsByRole`** (L346-361): JOIN `role_permission` 表
- **`GetPermissionsByResource`** (L363-378): 保持 LIKE（resources 是固定 ID 列表，量级小）

---

### Component 2: User 模型 IdP 字段清理

---

#### [MODIFY] [user.go](file:///Users/hs/workspace/github/casdoor/object/user.go)

1. **`UpdateUser` columns** (L776-790): 移除全部 84 个 IdP 列名
2. **`getUserByWechatId`** (L383-398): 改为查 `UserIdentity` 表 (providerType="wechat")

---

#### [MODIFY] [token_jwt.go](file:///Users/hs/workspace/github/casdoor/object/token_jwt.go)

1. **`UserWithoutThirdIdp`** struct (L66-154): 移除 12 个 IdP 字段 (GitHub/Google/QQ/WeChat/Facebook/DingTalk/Weibo/Gitee/LinkedIn/Wecom/Lark/Gitlab)
2. **`getUserWithoutThirdIdp`** 函数 (L220-310): 移除对应的 `user.GitHub` 等赋值行

---

#### [MODIFY] [user_test.go](file:///Users/hs/workspace/github/casdoor/object/user_test.go)

1. **`TestSyncAvatarsFromGitHub`** (L36-48): 改为从 `UserIdentity` 表查 GitHub providerId
2. **`TestGetUserByField`** (L105-114): DingTalk → 改为通过 `UserIdentity` 查

---

### Component 3: Role/Permission 写入同步

---

#### [MODIFY] [role.go](file:///Users/hs/workspace/github/casdoor/object/role.go)

- **`UpdateRole`**: 更新角色时同步更新 `user_role` 映射表
- **`AddRole`**: 新增角色后同步写入 `user_role`
- **`DeleteRole`**: 删除角色时清理 `user_role` 映射
- **`roleChangeTrigger`**: 角色改名时更新 `user_role`/`role_permission` 表中的 role 引用

---

#### [MODIFY] [permission.go](file:///Users/hs/workspace/github/casdoor/object/permission.go)

- **`UpdatePermission`**: 同步更新 `user_permission`/`role_permission`
- **`AddPermission`**: 同步写入映射表
- **`DeletePermission`**: 清理映射表

---

### Component 4: Docker MySQL + 编译验证

---

#### Docker MySQL

- 启动 MySQL 8.0 容器: `docker run -d --name casdoor-mysql -p 3306:3306 -e MYSQL_ROOT_PASSWORD=123456 -e MYSQL_DATABASE=casdoor mysql:8.0`
- 配置匹配现有 `app.conf`: `root:123456@tcp(localhost:3306)/casdoor`

---

## Verification Plan

### Automated Tests

1. **编译验证**: `cd /Users/hs/workspace/github/casdoor && go build ./...` -- 必须通过
2. **单元测试**: `go test ./object/... -v -run TestGetMaskedUsers` -- 不依赖 DB 的纯逻辑测试
3. **集成测试** (需 Docker MySQL):
   - `go test ./object/... -v -count=1` -- 全部 object 层测试

### Manual Verification

1. 启动 MySQL Docker + 启动 Casdoor: `go run main.go`
2. 验证 Sync2 新表创建成功 (检查 MySQL 中是否有 `user_identity`, `user_role`, `role_permission`, `user_permission` 表)
3. 验证迁移脚本正确执行 (如果已有数据)
