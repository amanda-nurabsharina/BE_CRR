package usecase

import (
	"context"
	"time"

	"be_crr/internal/domain"
	"be_crr/internal/domain/entity"
	"be_crr/internal/domain/repository"

	"github.com/google/uuid"
)

type AdminUsecase struct {
	tenantRepo repository.TenantRepository
	subRepo    repository.SubscriptionRepository
	auditRepo  repository.AuditRepository
}

func NewAdminUsecase(
	tenantRepo repository.TenantRepository,
	subRepo repository.SubscriptionRepository,
	auditRepo repository.AuditRepository,
) *AdminUsecase {
	return &AdminUsecase{
		tenantRepo: tenantRepo,
		subRepo:    subRepo,
		auditRepo:  auditRepo,
	}
}

type AdminMetricsDTO struct {
	TotalTenants          int64   `json:"total_tenants"`
	ActiveTenants         int64   `json:"active_tenants"`
	PendingPayments       int64   `json:"pending_payments"`
	TotalStorageUsedGB    float64 `json:"total_storage_used_gb"`
	TotalASRMinutesUsed   int     `json:"total_asr_minutes_used"`
}

func (u *AdminUsecase) GetMetrics(ctx context.Context) (*AdminMetricsDTO, error) {
	tenants, total, err := u.tenantRepo.ListAll(ctx, "", 1, 1000)
	if err != nil {
		return nil, err
	}

	var activeCount int64 = 0
	var totalStorageBytes int64 = 0
	var totalASRSeconds int = 0

	for _, t := range tenants {
		if t.Status == entity.TenantStatusActive {
			activeCount++
		}
		if t.Subscription != nil {
			totalStorageBytes += t.Subscription.UsedStorageBytes
			totalASRSeconds += t.Subscription.UsedTranscriptionSeconds
		}
	}

	_, pendingPaymentsCount, _ := u.subRepo.ListPayments(ctx, string(entity.PaymentStatusPending), 1, 10)

	return &AdminMetricsDTO{
		TotalTenants:        total,
		ActiveTenants:       activeCount,
		PendingPayments:     pendingPaymentsCount,
		TotalStorageUsedGB:  float64(totalStorageBytes) / (1024 * 1024 * 1024),
		TotalASRMinutesUsed: totalASRSeconds / 60,
	}, nil
}

func (u *AdminUsecase) ListTenants(ctx context.Context, status string, page, limit int) ([]entity.Tenant, int64, error) {
	return u.tenantRepo.ListAll(ctx, status, page, limit)
}

func (u *AdminUsecase) UpdateTenantStatus(ctx context.Context, tenantID uuid.UUID, status entity.TenantStatus) error {
	tenant, err := u.tenantRepo.GetByID(ctx, tenantID)
	if err != nil {
		return domain.ErrNotFound
	}
	tenant.Status = status
	tenant.UpdatedAt = time.Now()
	return u.tenantRepo.Update(ctx, tenant)
}

func (u *AdminUsecase) ListPendingPayments(ctx context.Context, page, limit int) ([]entity.ManualPayment, int64, error) {
	return u.subRepo.ListPayments(ctx, string(entity.PaymentStatusPending), page, limit)
}

func (u *AdminUsecase) ApprovePayment(ctx context.Context, adminID, paymentID uuid.UUID) error {
	payment, err := u.subRepo.GetPaymentByID(ctx, paymentID)
	if err != nil {
		return domain.ErrNotFound
	}

	now := time.Now()
	payment.Status = entity.PaymentStatusApproved
	payment.ReviewedByAdminID = &adminID
	payment.ReviewedAt = &now
	payment.UpdatedAt = now

	if err := u.subRepo.UpdatePayment(ctx, payment); err != nil {
		return err
	}

	// Activate Tenant & Subscription
	tenant, err := u.tenantRepo.GetByID(ctx, payment.TenantID)
	if err == nil {
		tenant.Status = entity.TenantStatusActive
		tenant.UpdatedAt = now
		_ = u.tenantRepo.Update(ctx, tenant)
	}

	sub, err := u.subRepo.GetByTenantID(ctx, payment.TenantID)
	if err == nil {
		sub.Status = entity.SubStatusActive
		newExpiry := now.AddDate(0, 1, 0)
		if sub.ExpiresAt != nil && sub.ExpiresAt.After(now) {
			newExpiry = sub.ExpiresAt.AddDate(0, 1, 0)
		}
		sub.ExpiresAt = &newExpiry
		sub.MaxTranscriptionSeconds += 18000 // +300 minutes
		sub.MaxStorageBytes += 21474836480    // +20 GB
		sub.UpdatedAt = now
		_ = u.subRepo.Update(ctx, sub)
	}

	return nil
}

func (u *AdminUsecase) RejectPayment(ctx context.Context, adminID, paymentID uuid.UUID, reason string) error {
	payment, err := u.subRepo.GetPaymentByID(ctx, paymentID)
	if err != nil {
		return domain.ErrNotFound
	}

	now := time.Now()
	payment.Status = entity.PaymentStatusRejected
	payment.ReviewedByAdminID = &adminID
	payment.ReviewedAt = &now
	payment.AdminNotes = reason
	payment.UpdatedAt = now

	return u.subRepo.UpdatePayment(ctx, payment)
}

func (u *AdminUsecase) AdjustQuota(ctx context.Context, tenantID uuid.UUID, addMinutes int, addStorageGB int) error {
	sub, err := u.subRepo.GetByTenantID(ctx, tenantID)
	if err != nil {
		return domain.ErrNotFound
	}

	sub.MaxTranscriptionSeconds += addMinutes * 60
	sub.MaxStorageBytes += int64(addStorageGB) * 1024 * 1024 * 1024
	sub.UpdatedAt = time.Now()

	return u.subRepo.Update(ctx, sub)
}

