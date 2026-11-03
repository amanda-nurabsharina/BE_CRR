package postgres

import (
	"context"

	"be_crr/internal/domain/entity"
	"be_crr/internal/domain/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TenantRepo struct {
	db *gorm.DB
}

func NewTenantRepo(db *gorm.DB) repository.TenantRepository {
	return &TenantRepo{db: db}
}

func (r *TenantRepo) Create(ctx context.Context, tenant *entity.Tenant) error {
	return r.db.WithContext(ctx).Create(tenant).Error
}

func (r *TenantRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Tenant, error) {
	var tenant entity.Tenant
	err := r.db.WithContext(ctx).Preload("Subscription").First(&tenant, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &tenant, nil
}

func (r *TenantRepo) GetByEmail(ctx context.Context, email string) (*entity.Tenant, error) {
	var tenant entity.Tenant
	err := r.db.WithContext(ctx).Preload("Subscription").First(&tenant, "owner_email = ?", email).Error
	if err != nil {
		return nil, err
	}
	return &tenant, nil
}

func (r *TenantRepo) ListAll(ctx context.Context, status string, page, limit int) ([]entity.Tenant, int64, error) {
	var tenants []entity.Tenant
	var total int64

	q := r.db.WithContext(ctx).Model(&entity.Tenant{})
	if status != "" {
		q = q.Where("status = ?", status)
	}

	q.Count(&total)

	offset := (page - 1) * limit
	err := q.Preload("Subscription").Offset(offset).Limit(limit).Order("created_at DESC").Find(&tenants).Error
	return tenants, total, err
}

func (r *TenantRepo) Update(ctx context.Context, tenant *entity.Tenant) error {
	return r.db.WithContext(ctx).Save(tenant).Error
}

type UserRepo struct {
	db *gorm.DB
}

func NewUserRepo(db *gorm.DB) repository.UserRepository {
	return &UserRepo{db: db}
}

func (r *UserRepo) Create(ctx context.Context, user *entity.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *UserRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).Preload("Tenant").Preload("Tenant.Subscription").First(&user, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).Preload("Tenant").Preload("Tenant.Subscription").First(&user, "email = ?", email).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepo) Update(ctx context.Context, user *entity.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

