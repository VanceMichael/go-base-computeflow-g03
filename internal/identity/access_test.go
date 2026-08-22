package identity_test

import (
	"errors"
	"github.com/VanceMichael/harborflow/internal/domain"
	"github.com/VanceMichael/harborflow/internal/identity"
	"testing"
)

func TestCoordinatorCanOperateAndAudit(t *testing.T) {
	if !identity.Can(domain.RoleCoordinator, identity.PermissionOperate) || !identity.Can(domain.RoleCoordinator, identity.PermissionAudit) {
		t.Fatal("coordinator permissions missing")
	}
}
func TestInspectorCannotAudit(t *testing.T) {
	if err := identity.Require(domain.RoleInspector, identity.PermissionAudit); !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("got %v", err)
	}
}
func TestAllowedRolesIncludesAuditor(t *testing.T) {
	roles := identity.AllowedRoles(identity.PermissionAudit)
	found := false
	for _, role := range roles {
		if role == domain.RoleAuditor {
			found = true
		}
	}
	if !found {
		t.Fatal("auditor missing")
	}
}
