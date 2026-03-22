# Task: casdoor-refactor

## 目标
执行 Route A 彻底重构，解决 RBAC N+1 性能雪崩以及 User 247 列的神对象问题。

## Checklists
- [x] 1. 模型定义
  - [x] `object/user_identity.go` 新增 UserIdentity 表
  - [x] `object/rbac_models.go` 新增 UserRole, RolePermission 等映射表
  - [x] `object/ormer.go` 注册新表
- [x] 2. 数据迁移清洗
  - [x] `object/migration.go` 编写单向不可逆数据迁移代码 (启动时执行)
- [x] 3. 核心对象瘦身与接口重写
  - [x] `object/user.go` 移除所有硬编码的数十个 IdP 字段
  - [x] `object/role.go` 移除 `LIKE` 改为联查
  - [x] `object/permission.go` 移除 `LIKE` 改为联查
- [x] 4. Adapter & Controller 适配
  - [x] `controllers/user.go` 还原 API 层面对外表现
  - [x] `controllers/role.go` 等权限管理接口适配新结构
- [x] 5. 测试验证
  - [x] `go test ./object/...`
