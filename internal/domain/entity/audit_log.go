package entity

import (
	"time"

	"github.com/google/uuid"
)

type AuditLog struct {
	ID           uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID     *uuid.UUID `json:"tenant_id" gorm:"type:uuid;index"`
	ActorID      uuid.UUID  `json:"actor_id" gorm:"type:uuid;not null"`
	Action       string     `json:"action" gorm:"type:varchar(100);not null"` // AUTH_LOGIN, RECORDING_INGEST, TRANSCRIPT_EXPORT, MANUAL_PURGE
	ResourceType string     `json:"resource_type" gorm:"type:varchar(100);not null"`
	ResourceID   uuid.UUID  `json:"resource_id" gorm:"type:uuid;not null"`
	IPAddress    string     `json:"ip_address" gorm:"type:varchar(45);not null"`
	UserAgent    string     `json:"user_agent" gorm:"type:text"`
	DetailsJSON  string     `json:"details_json" gorm:"type:text"`
	CreatedAt    time.Time  `json:"created_at" gorm:"not null;default:now()"`
}

