package entity

import (
	"time"

	"github.com/google/uuid"
)

type AudioSourceType string

const (
	SourceBrowserOnline  AudioSourceType = "browser_mic_online"
	SourceBrowserOffline AudioSourceType = "browser_mic_offline"
	SourceUploadAudio    AudioSourceType = "upload_audio"
	SourceExtractedMP4   AudioSourceType = "extracted_mp4"
)

type AudioEvidence struct {
	ID                  uuid.UUID       `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID            uuid.UUID       `json:"tenant_id" gorm:"type:uuid;index;not null"`
	CourtSessionID      uuid.UUID       `json:"court_session_id" gorm:"type:uuid;index;not null"`
	SourceType          AudioSourceType `json:"source_type" gorm:"type:varchar(50);not null"`
	OriginalFilename    string          `json:"original_filename" gorm:"type:varchar(255);not null"`
	StoragePath         string          `json:"storage_path" gorm:"type:text;not null"`
	FileSizeBytes       int64           `json:"file_size_bytes" gorm:"not null"`
	DurationSeconds     int             `json:"duration_seconds" gorm:"not null;default:0"`
	MimeType            string          `json:"mime_type" gorm:"type:varchar(100);not null"`
	Sha256Hash          string          `json:"sha256_hash" gorm:"type:char(64);not null"`
	EncryptionAlgorithm string          `json:"encryption_algorithm" gorm:"type:varchar(50);not null;default:'AES-256-GCM'"`
	IngestStatus        string          `json:"ingest_status" gorm:"type:varchar(50);not null;default:'ready'"` // ingesting, ready, failed
	CreatedAt           time.Time       `json:"created_at" gorm:"not null;default:now()"`

	Chunks      []AudioChunk `json:"chunks,omitempty" gorm:"foreignKey:AudioEvidenceID;references:ID;constraint:OnDelete:CASCADE"`
	Transcripts []Transcript `json:"transcripts,omitempty" gorm:"foreignKey:AudioEvidenceID;references:ID;constraint:OnDelete:CASCADE"`
}

type AudioChunk struct {
	ID              uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	AudioEvidenceID uuid.UUID `json:"audio_evidence_id" gorm:"type:uuid;index;not null"`
	SequenceNumber  int       `json:"sequence_number" gorm:"not null"`
	ChunkSizeBytes  int       `json:"chunk_size_bytes" gorm:"not null"`
	Crc32Checksum   string    `json:"crc32_checksum" gorm:"type:varchar(16);not null"`
	StoragePath     string    `json:"storage_path" gorm:"type:text;not null"`
	SyncedAt        time.Time `json:"synced_at" gorm:"not null;default:now()"`
}

