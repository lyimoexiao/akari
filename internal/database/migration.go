package database

import (
	"fmt"
	"net/http"

	"github.com/lyimoexiao/akari/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type permissionSeed struct {
	Role        string
	Name        string
	Object      string
	Action      string
	Description string
}

type userRoleMigration struct {
	Role string `gorm:"column:role"`
}

func (userRoleMigration) TableName() string {
	return "users"
}

var permissionSeeds = []permissionSeed{
	{model.RoleUser, "yggdrasil.status.read", "/api/v1/yggdrasil/user/status", http.MethodGet, "读取 Yggdrasil 用户状态"},
	{model.RoleUser, "skinlib.upload", "/api/v1/skinlib", http.MethodPost, "上传皮肤"},
	{model.RoleUser, "skinlib.update", "/api/v1/skinlib/:tid", http.MethodPut, "更新皮肤信息"},
	{model.RoleUser, "skinlib.delete", "/api/v1/skinlib/:tid", http.MethodDelete, "删除自己上传的皮肤"},
	{model.RoleUser, "closet.list", "/api/v1/closet", http.MethodGet, "查看衣橱"},
	{model.RoleUser, "closet.add", "/api/v1/closet", http.MethodPost, "添加皮肤到衣橱"},
	{model.RoleUser, "closet.remove", "/api/v1/closet/:tid", http.MethodDelete, "从衣橱移除皮肤"},
	{model.RoleUser, "closet.rename", "/api/v1/closet/:tid", http.MethodPut, "重命名衣橱物品"},
	{model.RoleStaff, "users.read", "/api/v1/users", http.MethodGet, "读取用户列表"},
	{model.RoleStaff, "users.verify-email", "/api/v1/users/verify-email", http.MethodPost, "验证用户邮箱"},
	{model.RoleStaff, "users.reset-password", "/api/v1/users/reset-password", http.MethodPost, "重置用户密码"},
	{model.RoleStaff, "users.delete", "/api/v1/users/:id", http.MethodDelete, "删除用户"},
	{model.RoleStaff, "skinlib.manage", "/api/v1/skinlib/manage", http.MethodGet, "管理皮肤库（查看私有/删除任意）"},
	{model.RoleSuperAdmin, "roles.read", "/api/v1/roles", http.MethodGet, "读取角色列表"},
	{model.RoleSuperAdmin, "roles.create", "/api/v1/roles", http.MethodPost, "创建角色"},
	{model.RoleSuperAdmin, "roles.update", "/api/v1/roles/:id", http.MethodPut, "更新角色"},
	{model.RoleSuperAdmin, "roles.delete", "/api/v1/roles/:id", http.MethodDelete, "删除角色"},
	{model.RoleSuperAdmin, "roles.set-default", "/api/v1/roles/:id/default", http.MethodPut, "设置默认注册角色"},
	{model.RoleSuperAdmin, "roles.assign", "/api/v1/roles/assign", http.MethodPost, "分配用户角色"},
	{model.RoleSuperAdmin, "permissions.read", "/api/v1/permissions", http.MethodGet, "读取权限列表"},
	{model.RoleSuperAdmin, "request-logs.read", "/api/v1/request-logs/:request_id", http.MethodGet, "按请求 ID 查询请求日志"},
}

var roleDescriptions = map[string]string{
	model.RoleSuperAdmin: "超级管理员",
	model.RoleStaff:      "STAFF",
	model.RoleUser:       "普通用户",
}

func Migrate(db *gorm.DB) error {
	hasLegacyRole := db.Migrator().HasColumn("users", "role")

	if err := db.AutoMigrate(
		&model.User{},
		&model.Role{},
		&model.Permission{},
		&model.UserRole{},
		&model.RolePermission{},
		&model.RequestLog{},
		&model.ScoreLog{},
		&model.Texture{},
		&model.UserCloset{},
	); err != nil {
		return fmt.Errorf("migrate access-control schema: %w", err)
	}
	if err := seedAccessControl(db); err != nil {
		return err
	}
	if err := migrateUserRoles(db, hasLegacyRole); err != nil {
		return err
	}
	if hasLegacyRole {
		if err := db.Migrator().DropColumn(&userRoleMigration{}, "Role"); err != nil {
			return fmt.Errorf("drop users.role: %w", err)
		}
	}
	if err := db.AutoMigrate(&model.User{}, &model.Role{}); err != nil {
		return fmt.Errorf("finalize access-control constraints: %w", err)
	}
	return nil
}

func seedAccessControl(db *gorm.DB) error {
	var roleCount, permissionCount int64
	if err := db.Model(&model.Role{}).Count(&roleCount).Error; err != nil {
		return fmt.Errorf("count roles: %w", err)
	}
	if err := db.Model(&model.Permission{}).Count(&permissionCount).Error; err != nil {
		return fmt.Errorf("count permissions: %w", err)
	}
	fresh := roleCount == 0 && permissionCount == 0
	seedNames := []string{model.RoleSuperAdmin}
	if fresh {
		seedNames = append(seedNames, model.RoleStaff, model.RoleUser)
	}
	for _, name := range seedNames {
		role := model.Role{Name: name, Description: roleDescriptions[name], IsDefault: fresh && name == model.RoleUser}
		if err := db.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "name"}}, DoNothing: true}).Create(&role).Error; err != nil {
			return fmt.Errorf("seed role %s: %w", name, err)
		}
	}
	roles := make(map[string]model.Role)
	var storedRoles []model.Role
	if err := db.Find(&storedRoles).Error; err != nil {
		return fmt.Errorf("load roles: %w", err)
	}
	for _, stored := range storedRoles {
		roles[stored.Name] = stored
	}
	var defaultCount int64
	if err := db.Model(&model.Role{}).Where("is_default = ?", true).Count(&defaultCount).Error; err != nil {
		return fmt.Errorf("count default roles: %w", err)
	}
	if defaultCount == 0 {
		if userRole, exists := roles[model.RoleUser]; exists {
			if err := db.Model(&userRole).Update("is_default", true).Error; err != nil {
				return fmt.Errorf("set registration role: %w", err)
			}
			userRole.IsDefault = true
			roles[model.RoleUser] = userRole
		}
	}
	if fresh {
		if err := setRoleParents(db, roles); err != nil {
			return err
		}
	}
	for _, seed := range permissionSeeds {
		permission := model.Permission{
			Name: seed.Name, Object: seed.Object, Action: seed.Action, Description: seed.Description,
		}
		result := db.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "name"}}, DoNothing: true}).Create(&permission)
		if result.Error != nil {
			return fmt.Errorf("seed permission %s: %w", seed.Name, result.Error)
		}
		if err := db.Where("name = ?", seed.Name).First(&permission).Error; err != nil {
			return fmt.Errorf("load permission %s: %w", seed.Name, err)
		}
		role, roleExists := roles[seed.Role]
		if result.RowsAffected > 0 && roleExists {
			relation := model.RolePermission{RoleID: role.ID, PermissionID: permission.ID}
			if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&relation).Error; err != nil {
				return fmt.Errorf("assign permission %s: %w", seed.Name, err)
			}
		}
	}
	return nil
}

func setRoleParents(db *gorm.DB, roles map[string]model.Role) error {
	parents := map[string]string{
		model.RoleStaff:      model.RoleUser,
		model.RoleSuperAdmin: model.RoleStaff,
	}
	for childName, parentName := range parents {
		if err := db.Model(&model.Role{}).
			Where("id = ?", roles[childName].ID).
			Update("parent_id", roles[parentName].ID).Error; err != nil {
			return fmt.Errorf("set parent for role %s: %w", childName, err)
		}
	}
	return nil
}

func migrateUserRoles(db *gorm.DB, hasLegacyRole bool) error {
	if !hasLegacyRole {
		return nil
	}

	roles := make(map[string]uint)
	var storedRoles []model.Role
	if err := db.Find(&storedRoles).Error; err != nil {
		return fmt.Errorf("load roles for user migration: %w", err)
	}
	for _, role := range storedRoles {
		roles[role.Name] = role.ID
	}

	var users []struct {
		ID   uint
		Role string
	}
	query := db.Table("users").Select("id, role")
	if err := query.Scan(&users).Error; err != nil {
		return fmt.Errorf("load users for role migration: %w", err)
	}
	for _, user := range users {
		roleName := user.Role
		if _, exists := roles[roleName]; !exists {
			roleName = model.RoleUser
		}
		roleID, exists := roles[roleName]
		if !exists {
			continue
		}
		var count int64
		if err := db.Model(&model.UserRole{}).Where("user_id = ?", user.ID).Count(&count).Error; err != nil {
			return fmt.Errorf("check roles for user %d: %w", user.ID, err)
		}
		if count > 0 {
			continue
		}
		assignment := model.UserRole{UserID: user.ID, RoleID: roleID}
		if err := db.Create(&assignment).Error; err != nil {
			return fmt.Errorf("migrate role for user %d: %w", user.ID, err)
		}
	}
	return nil
}
