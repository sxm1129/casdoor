package object

// syncRoleUserMappings rebuilds user_role mappings for a given role.
// It deletes existing mappings and recreates them from the role.Users slice.
func syncRoleUserMappings(role *Role) error {
	roleId := role.GetId()

	// Delete all existing mappings for this role
	_, err := ormer.Engine.Where("role = ?", roleId).Delete(&UserRole{})
	if err != nil {
		return err
	}

	// Insert new mappings from Users slice
	if len(role.Users) > 0 {
		mappings := make([]UserRole, 0, len(role.Users))
		for _, userId := range role.Users {
			mappings = append(mappings, UserRole{
				User: userId,
				Role: roleId,
			})
		}
		_, err = ormer.Engine.Insert(&mappings)
		if err != nil {
			return err
		}
	}
	return nil
}

// deleteRoleUserMappings removes all user_role mappings for a given role.
func deleteRoleUserMappings(roleId string) error {
	_, err := ormer.Engine.Where("role = ?", roleId).Delete(&UserRole{})
	return err
}

// syncPermissionUserMappings rebuilds user_permission mappings for a given permission.
func syncPermissionUserMappings(permission *Permission) error {
	permId := permission.GetId()

	// Delete existing user_permission mappings
	_, err := ormer.Engine.Where("permission = ?", permId).Delete(&UserPermission{})
	if err != nil {
		return err
	}

	// Insert from Users slice
	if len(permission.Users) > 0 {
		mappings := make([]UserPermission, 0, len(permission.Users))
		for _, userId := range permission.Users {
			mappings = append(mappings, UserPermission{
				User:       userId,
				Permission: permId,
			})
		}
		_, err = ormer.Engine.Insert(&mappings)
		if err != nil {
			return err
		}
	}
	return nil
}

// syncPermissionRoleMappings rebuilds role_permission mappings for a given permission.
func syncPermissionRoleMappings(permission *Permission) error {
	permId := permission.GetId()

	// Delete existing role_permission mappings
	_, err := ormer.Engine.Where("permission = ?", permId).Delete(&RolePermission{})
	if err != nil {
		return err
	}

	// Insert from Roles slice
	if len(permission.Roles) > 0 {
		mappings := make([]RolePermission, 0, len(permission.Roles))
		for _, roleId := range permission.Roles {
			mappings = append(mappings, RolePermission{
				Role:       roleId,
				Permission: permId,
			})
		}
		_, err = ormer.Engine.Insert(&mappings)
		if err != nil {
			return err
		}
	}
	return nil
}

// deletePermissionMappings removes all user_permission and role_permission mappings for a permission.
func deletePermissionMappings(permId string) error {
	_, err := ormer.Engine.Where("permission = ?", permId).Delete(&UserPermission{})
	if err != nil {
		return err
	}
	_, err = ormer.Engine.Where("permission = ?", permId).Delete(&RolePermission{})
	return err
}
