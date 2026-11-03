package entity

import (
	"time"

	"github.com/google/uuid"
)

type GenerationMode string

const (
	ModeAutoASR         GenerationMode = "auto_asr"
	ModeManualTyped     GenerationMode = "manual_typed"
	ModeHybridCorrected GenerationMode = "hybrid_corrected"
)

type Transcript struct {
	ID              uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID        uuid.UUID      `json:"tenant_id" gorm:"type:uuid;index;not null"`
	AudioEvidenceID uuid.UUID      `json:"audio_evidence_id" gorm:"type:uuid;index;not null"`
	GenerationMode  GenerationMode `json:"generation_mode" gorm:"type:varchar(50);not null"`
	Status          string         `json:"status" gorm:"type:varchar(50);not null;default:'draft'"` // processing_asr, draft, finalized
	VersionID       int            `json:"version_id" gorm:"not null;default:1"`
	CreatedAt       time.Time      `json:"created_at" gorm:"not null;default:now()"`
	UpdatedAt       time.Time      `json:"updated_at" gorm:"not null;default:now()"`

	Segments []TranscriptSegment `json:"segments,omitempty" gorm:"foreignKey:TranscriptID;references:ID;constraint:OnDelete:CASCADE"`
}

type TranscriptSegment struct {
	ID               uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TranscriptID     uuid.UUID `json:"transcript_id" gorm:"type:uuid;index;not null"`
	SpeakerTag       string    `json:"speaker_tag" gorm:"type:varchar(100);not null"` // Hakim Ketua, Penasihat Hukum, Jaksa, Saksi, Terdakwa
	StartTimeMs      int       `json:"start_time_ms" gorm:"not null"`
	EndTimeMs        int       `json:"end_time_ms" gorm:"not null"`
	TextContent      string    `json:"text_content" gorm:"type:text;not null"`
	SequenceOrder    int       `json:"sequence_order" gorm:"not null;index"`
	IsEditedManually bool      `json:"is_edited_manually" gorm:"default:false;not null"`
	UpdatedAt        time.Time `json:"updated_at" gorm:"not null;default:now()"`
}

