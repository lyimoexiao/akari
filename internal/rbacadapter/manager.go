package rbacadapter

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/casbin/casbin/v3"
	"github.com/lyimoexiao/akari/internal/role"
	"gorm.io/gorm"
)

type Manager struct {
	*RoleRepository
	db         *gorm.DB
	mutationMu sync.RWMutex
	enforcer   atomic.Pointer[casbin.SyncedEnforcer]
}

func NewManager(db *gorm.DB) (*Manager, error) {
	enforcer, err := buildEnforcer(db)
	if err != nil {
		return nil, err
	}
	manager := &Manager{db: db, RoleRepository: NewRoleRepository(db)}
	manager.enforcer.Store(enforcer)
	return manager, nil
}

func (manager *Manager) UpdatePolicy(ctx context.Context, change func(role.Repository) error) error {
	manager.mutationMu.Lock()
	defer manager.mutationMu.Unlock()

	tx := manager.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("begin policy transaction: %w", tx.Error)
	}
	if err := change(NewRoleRepository(tx)); err != nil {
		return rollback(tx, err)
	}
	candidate, err := buildEnforcer(tx)
	if err != nil {
		return rollback(tx, err)
	}
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("commit policy transaction: %w", err)
	}
	manager.enforcer.Store(candidate)
	return nil
}

func rollback(tx *gorm.DB, cause error) error {
	if err := tx.Rollback().Error; err != nil {
		return errors.Join(cause, fmt.Errorf("rollback policy transaction: %w", err))
	}
	return cause
}
