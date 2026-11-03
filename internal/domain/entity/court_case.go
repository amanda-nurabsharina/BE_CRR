package entity

import (
	"time"

	"github.com/google/uuid"
)

type CourtCase struct {
	ID               uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID         uuid.UUID `json:"tenant_id" gorm:"type:uuid;index;not null"`
	CaseNumber       string    `json:"case_number" gorm:"type:varchar(150);not null;index"` // e.g. 123/Pid.B/2026/PN JKT.SEL
	CourtName        string    `json:"court_name" gorm:"type:varchar(200);not null"`
	CaseCategory     string    `json:"case_category" gorm:"type:varchar(100);not null"` // Pidana Umum, Perdata, Niaga, TUN
	ClientName       string    `json:"client_name" gorm:"type:varchar(255);not null"`
	CounterPartyName string    `json:"counter_party_name" gorm:"type:varchar(255)"`
	JudgePanelNames  string    `json:"judge_panel_names" gorm:"type:text"`
	ProsecutorNames  string    `json:"prosecutor_names" gorm:"type:text"`
	ClerkName        string    `json:"clerk_name" gorm:"type:varchar(255)"`
	Notes            string    `json:"notes" gorm:"type:text"`
	CreatedAt        time.Time `json:"created_at" gorm:"not null;default:now()"`
	UpdatedAt        time.Time `json:"updated_at" gorm:"not null;default:now()"`

	Sessions []CourtSession `json:"sessions,omitempty" gorm:"foreignKey:CourtCaseID;references:ID;constraint:OnDelete:CASCADE"`
}

type CourtSessionStatus string

const (
	SessionStatusScheduled CourtSessionStatus = "scheduled"
	SessionStatusRecording CourtSessionStatus = "recording"
	SessionStatusCompleted CourtSessionStatus = "completed"
)

type CourtSession struct {
	ID           uuid.UUID          `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID     uuid.UUID          `json:"tenant_id" gorm:"type:uuid;index;not null"`
	CourtCaseID  uuid.UUID          `json:"court_case_id" gorm:"type:uuid;index;not null"`
	SessionTitle string             `json:"session_title" gorm:"type:varchar(255);not null"` // e.g. Agenda: Pembuktian Saksi A
	SessionDate  time.Time          `json:"session_date" gorm:"type:date;not null"`
	StartTime    *time.Time         `json:"start_time"`
	EndTime      *time.Time         `json:"end_time"`
	Status       CourtSessionStatus `json:"status" gorm:"type:varchar(50);not null;default:'scheduled'"`
	CreatedAt    time.Time          `json:"created_at" gorm:"not null;default:now()"`
	UpdatedAt    time.Time          `json:"updated_at" gorm:"not null;default:now()"`

	AudioEvidences []AudioEvidence `json:"audio_evidences,omitempty" gorm:"foreignKey:CourtSessionID;references:ID;constraint:OnDelete:CASCADE"`
	Reports        []CourtReport   `json:"reports,omitempty" gorm:"foreignKey:CourtSessionID;references:ID;constraint:OnDelete:CASCADE"`
}
