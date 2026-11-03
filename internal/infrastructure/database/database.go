package database

import (
	"fmt"
	"log"
	"time"

	"be_crr/internal/domain/entity"
	"be_crr/internal/infrastructure/config"
	"be_crr/internal/infrastructure/security"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func InitDB(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=Asia/Jakarta",
		cfg.DBHost, cfg.DBUser, cfg.DBPass, cfg.DBName, cfg.DBPort, cfg.DBSSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Run auto migrations
	err = db.AutoMigrate(
		&entity.Tenant{},
		&entity.User{},
		&entity.Subscription{},
		&entity.ManualPayment{},
		&entity.CourtCase{},
		&entity.CourtSession{},
		&entity.AudioEvidence{},
		&entity.AudioChunk{},
		&entity.Transcript{},
		&entity.TranscriptSegment{},
		&entity.CourtReport{},
		&entity.ReportSignature{},
		&entity.AuditLog{},
	)
	if err != nil {
		return nil, fmt.Errorf("auto migration failed: %w", err)
	}

	// Seed Initial Data
	seedInitialData(db)

	return db, nil
}

func seedInitialData(db *gorm.DB) {
	// 1. Seed Super Admin
	var adminCount int64
	db.Model(&entity.User{}).Where("role = ?", entity.RoleSuperAdmin).Count(&adminCount)
	if adminCount == 0 {
		hash, _ := security.HashPassword("admin12345")
		admin := entity.User{
			FullName:     "Super Admin Platform",
			Email:        "admin@court-record.id",
			PasswordHash: hash,
			Role:         entity.RoleSuperAdmin,
			PhoneNumber:  "081234567890",
			IsActive:     true,
		}
		if err := db.Create(&admin).Error; err == nil {
			log.Println("[SEED] Created Super Admin: admin@court-record.id / admin12345")
		}
	}

	// 2. Seed Demo Tenant & Lawyer
	var tenantCount int64
	db.Model(&entity.Tenant{}).Count(&tenantCount)
	if tenantCount == 0 {
		tenantID := uuid.New()
		tenant := entity.Tenant{
			ID:          tenantID,
			Name:        "Budi Santoso & Partners Legal Office",
			OwnerEmail:  "advokat.budi@hukum.id",
			PhoneNumber: "081987654321",
			Status:      entity.TenantStatusActive,
		}
		db.Create(&tenant)

		// Create Subscription
		now := time.Now()
		expires := now.AddDate(0, 1, 0)
		sub := entity.Subscription{
			TenantID:                 tenantID,
			TierName:                 "pro_solo",
			Status:                   entity.SubStatusActive,
			MaxStorageBytes:          21474836480, // 20 GB
			UsedStorageBytes:         524288000,   // ~500 MB demo
			MaxTranscriptionSeconds:  18000,       // 300 minutes
			UsedTranscriptionSeconds: 3600,        // 60 minutes used
			StartedAt:                &now,
			ExpiresAt:                &expires,
		}
		db.Create(&sub)

		// Create Lawyer User
		hash, _ := security.HashPassword("admin12345")
		lawyer := entity.User{
			TenantID:     &tenantID,
			FullName:     "Budi Santoso, S.H., M.H.",
			Email:        "advokat.budi@hukum.id",
			PasswordHash: hash,
			Role:         entity.RoleTenantOwner,
			PhoneNumber:  "081987654321",
			BarNumber:    "PERADI/2020/09871",
			IsActive:     true,
		}
		db.Create(&lawyer)

		// Create Demo Court Case
		caseID := uuid.New()
		courtCase := entity.CourtCase{
			ID:               caseID,
			TenantID:         tenantID,
			CaseNumber:       "89/Pid.B/2026/PN.JKT.PST",
			CourtName:        "Pengadilan Negeri Jakarta Pusat",
			CaseCategory:     "Pidana Umum",
			ClientName:       "Ahmad Fauzi (Terdakwa)",
			CounterPartyName: "Jaksa Penuntut Umum Kejari Jakpus",
			JudgePanelNames:  "Dr. Wahyu Iman S.H. (Ketua), Hendra S.H., Siti S.H.",
			ProsecutorNames:  "Bambang Suryo, S.H.",
			ClerkName:        "Dewi Sartika, S.H.",
			Notes:            "Perkara dugaan pelanggaran pasal 378 KUHP.",
		}
		db.Create(&courtCase)

		// Create Court Session
		sessID := uuid.New()
		session := entity.CourtSession{
			ID:           sessID,
			TenantID:     tenantID,
			CourtCaseID:  caseID,
			SessionTitle: "Sidang Pembuktian: Pemeriksaan Saksi Ahli Pidana",
			SessionDate:  time.Now(),
			Status:       entity.SessionStatusScheduled,
		}
		db.Create(&session)

		log.Println("[SEED] Created Demo Tenant & Case: advokat.budi@hukum.id / admin12345")
	}
}

