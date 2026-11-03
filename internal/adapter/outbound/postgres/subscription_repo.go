package postgres

import (
	"context"

	"be_crr/internal/domain/entity"
	"be_crr/internal/domain/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SubscriptionRepo struct {
	db *gorm.DB
}

func NewSubscriptionRepo(db *gorm.DB) repository.SubscriptionRepository {
	return &SubscriptionRepo{db: db}
}

func (r *SubscriptionRepo) Create(ctx context.Context, sub *entity.Subscription) error {
	return r.db.WithContext(ctx).Create(sub).Error
}

func (r *SubscriptionRepo) GetByTenantID(ctx context.Context, tenantID uuid.UUID) (*entity.Subscription, error) {
	var sub entity.Subscription
	err := r.db.WithContext(ctx).First(&sub, "tenant_id = ?", tenantID).Error
	if err != nil {
		return nil, err
	}
	return &sub, nil
}

func (r *SubscriptionRepo) Update(ctx context.Context, sub *entity.Subscription) error {
	return r.db.WithContext(ctx).Save(sub).Error
}

func (r *SubscriptionRepo) CreatePayment(ctx context.Context, payment *entity.ManualPayment) error {
	return r.db.WithContext(ctx).Create(payment).Error
}

func (r *SubscriptionRepo) GetPaymentByID(ctx context.Context, id uuid.UUID) (*entity.ManualPayment, error) {
	var payment entity.ManualPayment
	err := r.db.WithContext(ctx).Preload("Tenant").First(&payment, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &payment, nil
}

func (r *SubscriptionRepo) ListPayments(ctx context.Context, status string, page, limit int) ([]entity.ManualPayment, int64, error) {
	var payments []entity.ManualPayment
	var total int64

	q := r.db.WithContext(ctx).Model(&entity.ManualPayment{})
	if status != "" {
		q = q.Where("status = ?", status)
	}

	q.Count(&total)

	offset := (page - 1) * limit
	err := q.Preload("Tenant").Offset(offset).Limit(limit).Order("created_at DESC").Find(&payments).Error
	return payments, total, err
}

func (r *SubscriptionRepo) UpdatePayment(ctx context.Context, payment *entity.ManualPayment) error {
	return r.db.WithContext(ctx).Save(payment).Error
}

