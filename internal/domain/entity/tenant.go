package entity

import (
	"time"

	"github.com/google/uuid"
)

type TenantStatus string

const (
	TenantStatusPendingActivation TenantStatus = "pending_activation"
	TenantStatusActive            TenantStatus = "active"
	TenantStatusSuspended         TenantStatus = "suspended"
	TenantStatusExpired           TenantStatus = "expired"
)

type Tenant struct {
	ID          uuid.UUID    `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name        string       `json:"name" gorm:"type:varchar(255);not null"`
	OwnerEmail  string       `json:"owner_email" gorm:"type:varchar(255);uniqueIndex;not null"`
	PhoneNumber string       `json:"phone_number" gorm:"type:varchar(50);not null"`
	Status      TenantStatus `json:"status" gorm:"type:varchar(50);not null;default:'pending_activation'"`
	CreatedAt   time.Time    `json:"created_at" gorm:"not null;default:now()"`
	UpdatedAt   time.Time    `json:"updated_at" gorm:"not null;default:now()"`

	Subscription *Subscription `json:"subscription,omitempty" gorm:"foreignKey:TenantID;references:ID"`
}

type SubscriptionStatus string

const (
	SubStatusPendingApproval SubscriptionStatus = "pending_approval"
	SubStatusActive          SubscriptionStatus = "active"
	SubStatusGracePeriod     SubscriptionStatus = "grace_period"
	SubStatusExpired         SubscriptionStatus = "expired"
)

type Subscription struct {
	ID                       uuid.UUID          `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID                 uuid.UUID          `json:"tenant_id" gorm:"type:uuid;index;not null"`
	TierName                 string             `json:"tier_name" gorm:"type:varchar(50);not null;default:'pro_solo'"`
	Status                   SubscriptionStatus `json:"status" gorm:"type:varchar(50);not null;default:'pending_approval'"`
	MaxStorageBytes          int64              `json:"max_storage_bytes" gorm:"not null;default:21474836480"` // 20GB
	UsedStorageBytes         int64              `json:"used_storage_bytes" gorm:"not null;default:0"`
	MaxTranscriptionSeconds  int                `json:"max_transcription_seconds" gorm:"not null;default:18000"` // 300 minutes
	UsedTranscriptionSeconds int                `json:"used_transcription_seconds" gorm:"not null;default:0"`
	StartedAt                *time.Time         `json:"started_at"`
	ExpiresAt                *time.Time         `json:"expires_at"`
	CreatedAt                time.Time          `json:"created_at" gorm:"not null;default:now()"`
	UpdatedAt                time.Time          `json:"updated_at" gorm:"not null;default:now()"`
}

type PaymentStatus string

const (
	PaymentStatusPending  PaymentStatus = "pending_review"
	PaymentStatusApproved PaymentStatus = "approved"
	PaymentStatusRejected PaymentStatus = "rejected"
)

type ManualPayment struct {
	ID                    uuid.UUID     `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID              uuid.UUID     `json:"tenant_id" gorm:"type:uuid;index;not null"`
	InvoiceNumber         string        `json:"invoice_number" gorm:"type:varchar(100);uniqueIndex;not null"`
	Amount                float64       `json:"amount" gorm:"type:numeric(14,2);not null"`
	BankDestination       string        `json:"bank_destination" gorm:"type:varchar(100);not null"`
	BankSenderName        string        `json:"bank_sender_name" gorm:"type:varchar(255);not null"`
	TransferDate          string        `json:"transfer_date" gorm:"type:varchar(50);not null"`
	ProofImageStoragePath string        `json:"proof_image_storage_path" gorm:"type:text;not null"`
	Status                PaymentStatus `json:"status" gorm:"type:varchar(50);not null;default:'pending_review'"`
	ReviewedByAdminID     *uuid.UUID    `json:"reviewed_by_admin_id"`
	ReviewedAt            *time.Time    `json:"reviewed_at"`
	AdminNotes            string        `json:"admin_notes" gorm:"type:text"`
	CreatedAt             time.Time     `json:"created_at" gorm:"not null;default:now()"`
	UpdatedAt             time.Time     `json:"updated_at" gorm:"not null;default:now()"`

	Tenant *Tenant `json:"tenant,omitempty" gorm:"foreignKey:TenantID;references:ID"`
}

