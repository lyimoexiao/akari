package rbacadapter

import (
	"fmt"

	"github.com/lyimoexiao/akari/internal/model"
	"github.com/lyimoexiao/akari/internal/permission"
)

func (manager *Manager) Snapshot() (permission.Snapshot, error) {
	manager.mutationMu.RLock()
	defer manager.mutationMu.RUnlock()

	policies, err := manager.enforcer.Load().GetPolicy()
	if err != nil {
		return permission.Snapshot{}, fmt.Errorf("get casbin policies: %w", err)
	}
	grouping, err := manager.enforcer.Load().GetGroupingPolicy()
	if err != nil {
		return permission.Snapshot{}, fmt.Errorf("get casbin role inheritance: %w", err)
	}
	rules := make([]permission.Rule, 0, len(policies))
	for _, policy := range policies {
		rules = append(rules, permission.Rule{Role: policy[0], Object: policy[1], Action: policy[2]})
	}
	inheritance := make([]permission.RoleInheritance, 0, len(grouping))
	for _, relation := range grouping {
		inheritance = append(inheritance, permission.RoleInheritance{Role: relation[0], Parent: relation[1]})
	}
	var roles []model.Role
	if err := manager.db.Order("id").Find(&roles).Error; err != nil {
		return permission.Snapshot{}, fmt.Errorf("list roles: %w", err)
	}
	roleNames := make([]string, len(roles))
	for index := range roles {
		roleNames[index] = roles[index].Name
	}
	return permission.Snapshot{Roles: roleNames, Rules: rules, Inheritance: inheritance}, nil
}
