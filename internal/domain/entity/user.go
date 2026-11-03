package entity

import (
	"time"

	"github.com/google/uuid"
)

type UserRole string

const (
	RoleSuperAdmin  UserRole = "SUPER_ADMIN"
	RoleTenantOwner UserRole = "TENANT_OWNER"
)

type User struct {
	ID           uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID     *uuid.UUID `json:"tenant_id" gorm:"type:uuid;index"` // null for Super Admin
	FullName     string     `json:"full_name" gorm:"type:varchar(255);not null"`
	Email        string     `json:"email" gorm:"type:varchar(255);uniqueIndex;not null"`
	PasswordHash string     `json:"-" gorm:"type:varchar(255);not null"`
	Role         UserRole   `json:"role" gorm:"type:varchar(50);not null;default:'TENANT_OWNER'"`
	PhoneNumber  string     `json:"phone_number" gorm:"type:varchar(50)"`
	BarNumber    string     `json:"bar_number" gorm:"type:varchar(100)"` // Nomor Induk Advokat (NIA)
	IsActive     bool       `json:"is_active" gorm:"default:true;not null"`
	CreatedAt    time.Time  `json:"created_at" gorm:"not null;default:now()"`
	UpdatedAt    time.Time  `json:"updated_at" gorm:"not null;default:now()"`

	Tenant *Tenant `json:"tenant,omitempty" gorm:"foreignKey:TenantID;references:ID"`
}

