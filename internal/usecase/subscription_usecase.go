package usecase

import (
	"context"
	"fmt"
	"io"
	"time"

	"be_crr/internal/domain"
	"be_crr/internal/domain/entity"
	"be_crr/internal/domain/repository"
	"be_crr/internal/infrastructure/storage"

	"github.com/google/uuid"
)

type SubscriptionUsecase struct {
	subRepo repository.SubscriptionRepository
	storage storage.Storage
}

func NewSubscriptionUsecase(subRepo repository.SubscriptionRepository, storage storage.Storage) *SubscriptionUsecase {
	return &SubscriptionUsecase{
		subRepo: subRepo,
		storage: storage,
	}
}

type SubscriptionUsageDTO struct {
	Subscription *entity.Subscription `json:"subscription"`
	StorageUsedPercent float64        `json:"storage_used_percent"`
	ASRUsedPercent     float64        `json:"asr_used_percent"`
}

func (u *SubscriptionUsecase) GetCurrentSubscription(ctx context.Context, tenantID uuid.UUID) (*SubscriptionUsageDTO, error) {
	sub, err := u.subRepo.GetByTenantID(ctx, tenantID)
	if err != nil {
		return nil, domain.ErrNotFound
	}

	storagePercent := 0.0
	if sub.MaxStorageBytes > 0 {
		storagePercent = (float64(sub.UsedStorageBytes) / float64(sub.MaxStorageBytes)) * 100.0
	}

	asrPercent := 0.0
	if sub.MaxTranscriptionSeconds > 0 {
		asrPercent = (float64(sub.UsedTranscriptionSeconds) / float64(sub.MaxTranscriptionSeconds)) * 100.0
	}

	return &SubscriptionUsageDTO{
		Subscription:       sub,
		StorageUsedPercent: storagePercent,
		ASRUsedPercent:     asrPercent,
	}, nil
}

type SubmitPaymentDTO struct {
	TenantID        uuid.UUID
	Amount          float64
	BankDestination string
	BankSenderName  string
	TransferDate    string
	ProofFilename   string
	ProofReader     io.Reader
}

func (u *SubscriptionUsecase) SubmitManualPayment(ctx context.Context, dto SubmitPaymentDTO) (*entity.ManualPayment, error) {
	subDir := fmt.Sprintf("receipts/%s", dto.TenantID.String())
	filename := fmt.Sprintf("receipt_%d_%s", time.Now().Unix(), dto.ProofFilename)
	savedPath, _, err := u.storage.SaveFile(subDir, filename, dto.ProofReader)
	if err != nil {
		return nil, fmt.Errorf("failed to save transfer proof: %w", err)
	}

	payment := &entity.ManualPayment{
		TenantID:              dto.TenantID,
		InvoiceNumber:         fmt.Sprintf("INV-%s-%d", time.Now().Format("20060102"), time.Now().Unix()%10000),
		Amount:                dto.Amount,
		BankDestination:       dto.BankDestination,
		BankSenderName:        dto.BankSenderName,
		TransferDate:          dto.TransferDate,
		ProofImageStoragePath: savedPath,
		Status:                entity.PaymentStatusPending,
	}

	if err := u.subRepo.CreatePayment(ctx, payment); err != nil {
		return nil, err
	}

	return payment, nil
}

func (u *SubscriptionUsecase) GetPaymentHistory(ctx context.Context, tenantID uuid.UUID) ([]entity.ManualPayment, error) {
	payments, _, err := u.subRepo.ListPayments(ctx, "", 1, 100)
	if err != nil {
		return nil, err
	}

	var tenantPayments []entity.ManualPayment
	for _, p := range payments {
		if p.TenantID == tenantID {
			tenantPayments = append(tenantPayments, p)
		}
	}
	return tenantPayments, nil
}

