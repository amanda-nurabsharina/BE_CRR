package entity

import (
	"time"

	"github.com/google/uuid"
)

type CourtReportStatus string

const (
	ReportStatusDraft     CourtReportStatus = "draft"
	ReportStatusSigned    CourtReportStatus = "signed"
	ReportStatusFinalized CourtReportStatus = "finalized"
	ReportStatusPurged    CourtReportStatus = "purged"
)

type CourtReport struct {
	ID                 uuid.UUID         `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID           uuid.UUID         `json:"tenant_id" gorm:"type:uuid;index;not null"`
	CourtSessionID     uuid.UUID         `json:"court_session_id" gorm:"type:uuid;index;not null"`
	ReportTitle        string            `json:"report_title" gorm:"type:varchar(255);not null;default:'Berita Acara Jalannya Sidang'"`
	DocumentNumber     string            `json:"document_number" gorm:"type:varchar(150)"`
	CaseSummary        string            `json:"case_summary" gorm:"type:text"`
	KeyTestimoniesJSON string            `json:"key_testimonies_json" gorm:"type:text;default:'[]'"`
	ObjectionsNotes    string            `json:"objections_notes" gorm:"type:text"`
	PdfStoragePath     string            `json:"pdf_storage_path" gorm:"type:text"`
	PdfSha256Hash      string            `json:"pdf_sha256_hash" gorm:"type:char(64)"`
	Status             CourtReportStatus `json:"status" gorm:"type:varchar(50);not null;default:'draft'"`
	Version            int               `json:"version" gorm:"not null;default:1"`
	CreatedAt          time.Time         `json:"created_at" gorm:"not null;default:now()"`
	UpdatedAt          time.Time         `json:"updated_at" gorm:"not null;default:now()"`

	Session    *CourtSession    `json:"session,omitempty" gorm:"foreignKey:CourtSessionID;references:ID"`
	Signatures []ReportSignature `json:"signatures,omitempty" gorm:"foreignKey:CourtReportID;references:ID;constraint:OnDelete:CASCADE"`
}

type ReportSignature struct {
	ID                 uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CourtReportID      uuid.UUID `json:"court_report_id" gorm:"type:uuid;index;not null"`
	VerificationToken  string    `json:"verification_token" gorm:"type:varchar(100);uniqueIndex;not null"`
	SignerName         string    `json:"signer_name" gorm:"type:varchar(255);not null"`
	SignerIdentifier   string    `json:"signer_identifier" gorm:"type:varchar(100)"` // NIK / NIA
	SignerRole         string    `json:"signer_role" gorm:"type:varchar(100);not null"`
	SignatureImagePath string    `json:"signature_image_path" gorm:"type:text;not null"`
	SourceAudioSha256  string    `json:"source_audio_sha256" gorm:"type:char(64);not null"`
	DocumentSha256     string    `json:"document_sha256" gorm:"type:char(64);not null"`
	SignerIPAddress    string    `json:"signer_ip_address" gorm:"type:varchar(45);not null"`
	SignedAt           time.Time `json:"signed_at" gorm:"not null;default:now()"`
}

