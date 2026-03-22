package object

// AUDIT SMELL-03 fix: added unique composite indexes to prevent duplicate mappings

type UserRole struct {
	Id   int    `xorm:"int notnull pk autoincr" json:"id"`
	User string `xorm:"varchar(200) notnull unique(user_role)" json:"user"`
	Role string `xorm:"varchar(200) notnull unique(user_role)" json:"role"`
}

type RolePermission struct {
	Id         int    `xorm:"int notnull pk autoincr" json:"id"`
	Role       string `xorm:"varchar(200) notnull unique(role_perm)" json:"role"`
	Permission string `xorm:"varchar(200) notnull unique(role_perm)" json:"permission"`
}

type UserPermission struct {
	Id         int    `xorm:"int notnull pk autoincr" json:"id"`
	User       string `xorm:"varchar(200) notnull unique(user_perm)" json:"user"`
	Permission string `xorm:"varchar(200) notnull unique(user_perm)" json:"permission"`
}
