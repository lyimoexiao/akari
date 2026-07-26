package rbacadapter

import (
	"fmt"

	"github.com/casbin/casbin/v3"
	casbinmodel "github.com/casbin/casbin/v3/model"
	"github.com/lyimoexiao/akari/internal/model"
	"github.com/lyimoexiao/akari/internal/permission"
	"gorm.io/gorm"
)

const accessModelText = `[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && keyMatch2(r.obj, p.obj) && r.act == p.act
`

func buildEnforcer(db *gorm.DB) (*casbin.SyncedEnforcer, error) {
	accessModel, err := casbinmodel.NewModelFromString(accessModelText)
	if err != nil {
		return nil, fmt.Errorf("parse casbin model: %w", err)
	}
	enforcer, err := casbin.NewSyncedEnforcer(accessModel)
	if err != nil {
		return nil, fmt.Errorf("create casbin enforcer: %w", err)
	}
	if err := loadPolicy(db, enforcer); err != nil {
		return nil, err
	}
	return enforcer, nil
}

func loadPolicy(db *gorm.DB, enforcer *casbin.SyncedEnforcer) error {
	var rules []permission.Rule
	if err := db.Table("role_permissions").
		Select("roles.name AS role, permissions.object, permissions.action").
		Joins("JOIN roles ON roles.id = role_permissions.role_id").
		Joins("JOIN permissions ON permissions.id = role_permissions.permission_id").
		Scan(&rules).Error; err != nil {
		return fmt.Errorf("load permissions: %w", err)
	}
	policies := make([][]string, len(rules))
	for index := range rules {
		policies[index] = []string{rules[index].Role, rules[index].Object, rules[index].Action}
	}
	if _, err := enforcer.AddPoliciesEx(policies); err != nil {
		return fmt.Errorf("load casbin permissions: %w", err)
	}

	var roles []model.Role
	if err := db.Preload("Parent").Find(&roles).Error; err != nil {
		return fmt.Errorf("load role hierarchy: %w", err)
	}
	grouping := make([][]string, 0, len(roles))
	for _, storedRole := range roles {
		if storedRole.Parent != nil {
			grouping = append(grouping, []string{storedRole.Name, storedRole.Parent.Name})
		}
	}
	if _, err := enforcer.AddGroupingPoliciesEx(grouping); err != nil {
		return fmt.Errorf("load casbin role inheritance: %w", err)
	}
	return nil
}
