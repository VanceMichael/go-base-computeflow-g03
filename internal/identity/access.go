package identity

import (
	"fmt"
	"github.com/VanceMichael/harborflow/internal/domain"
)

type Permission string

const (
	PermissionOperate  Permission = "operate"
	PermissionInspect  Permission = "inspect"
	PermissionDispatch Permission = "dispatch"
	PermissionAudit    Permission = "audit"
)

var rolePermissions = map[domain.Role]map[Permission]bool{domain.RoleCoordinator: {PermissionOperate: true, PermissionInspect: true, PermissionDispatch: true, PermissionAudit: true}, domain.RoleInspector: {PermissionInspect: true}, domain.RoleDispatcher: {PermissionDispatch: true, PermissionInspect: true}, domain.RoleAuditor: {PermissionAudit: true}}

func Can(role domain.Role, permission Permission) bool { return rolePermissions[role][permission] }
func Require(role domain.Role, permission Permission) error {
	if !Can(role, permission) {
		return fmt.Errorf("%w: %s requires %s", domain.ErrUnauthorized, role, permission)
	}
	return nil
}
func AllowedRoles(permission Permission) []domain.Role {
	var out []domain.Role
	for role, permissions := range rolePermissions {
		if permissions[permission] {
			out = append(out, role)
		}
	}
	return out
}
