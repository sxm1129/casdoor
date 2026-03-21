package object

type UserRole struct {
	Id       int    `xorm:"int notnull pk autoincr" json:"id"`
	User     string `xorm:"varchar(200) notnull index" json:"user"`
	Role     string `xorm:"varchar(200) notnull index" json:"role"`
}

type RolePermission struct {
	Id         int    `xorm:"int notnull pk autoincr" json:"id"`
	Role       string `xorm:"varchar(200) notnull index" json:"role"`
	Permission string `xorm:"varchar(200) notnull index" json:"permission"`
}

type UserPermission struct {
	Id         int    `xorm:"int notnull pk autoincr" json:"id"`
	User       string `xorm:"varchar(200) notnull index" json:"user"`
	Permission string `xorm:"varchar(200) notnull index" json:"permission"`
}
