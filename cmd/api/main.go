package main

import (
	"log"

	"be_crr/internal/adapter/inbound/http"
	"be_crr/internal/adapter/inbound/http/handler"
	"be_crr/internal/adapter/outbound/postgres"
	"be_crr/internal/infrastructure/config"
	"be_crr/internal/infrastructure/database"
	"be_crr/internal/infrastructure/storage"
	"be_crr/internal/usecase"

	"github.com/gofiber/fiber/v2"
)

func main() {
	log.Println("[INFO] Starting Court Recording Report API Server...")

	// 1. Config
	cfg := config.LoadConfig()

	// 2. Database & Migrations
	db, err := database.InitDB(cfg)
	if err != nil {
		log.Fatalf("[FATAL] Database connection failed: %v", err)
	}
	log.Println("[INFO] Database connected & auto-migrations successfully applied.")

	// 3. Storage
	store, err := storage.NewLocalStorage(cfg.StorageDir)
	if err != nil {
		log.Fatalf("[FATAL] Local storage init failed: %v", err)
	}

	// 4. Repositories
	tenantRepo := postgres.NewTenantRepo(db)
	userRepo := postgres.NewUserRepo(db)
	subRepo := postgres.NewSubscriptionRepo(db)
	caseRepo := postgres.NewCaseRepo(db)
	sessionRepo := postgres.NewSessionRepo(db)
	audioRepo := postgres.NewAudioRepo(db)
	transRepo := postgres.NewTranscriptRepo(db)
	reportRepo := postgres.NewReportRepo(db)
	auditRepo := postgres.NewAuditRepo(db)

	// 5. Use Cases
	authUsecase := usecase.NewAuthUsecase(userRepo, tenantRepo, subRepo, cfg.JWTSecret)
	adminUsecase := usecase.NewAdminUsecase(tenantRepo, subRepo, auditRepo)
	subUsecase := usecase.NewSubscriptionUsecase(subRepo, store)
	caseUsecase := usecase.NewCaseUsecase(caseRepo, sessionRepo, audioRepo, subRepo, store, auditRepo)
	recUsecase := usecase.NewRecordingUsecase(audioRepo, sessionRepo, subRepo, store)
	ingestUsecase := usecase.NewIngestUsecase(audioRepo, sessionRepo, subRepo, store)
	transUsecase := usecase.NewTranscriptUsecase(transRepo, audioRepo, subRepo)
	reportUsecase := usecase.NewReportUsecase(reportRepo, sessionRepo, audioRepo, store, auditRepo)

	// 6. HTTP Handlers
	authHandler := handler.NewAuthHandler(authUsecase)
	adminHandler := handler.NewAdminHandler(adminUsecase)
	subHandler := handler.NewSubscriptionHandler(subUsecase)
	caseHandler := handler.NewCaseHandler(caseUsecase)
	recHandler := handler.NewRecordingHandler(recUsecase, ingestUsecase)
	transHandler := handler.NewTranscriptHandler(transUsecase)
	reportHandler := handler.NewReportHandler(reportUsecase)

	// 7. Fiber App & Router
	app := fiber.New(fiber.Config{
		AppName:      "Court Recording Report API v3.0",
		BodyLimit:    500 * 1024 * 1024, // 500 MB for media upload
		ServerHeader: "Court-Recording-Engine",
	})

	http.SetupRouter(app, http.RouterDependencies{
		JWTSecret:     cfg.JWTSecret,
		AuthHandler:   authHandler,
		AdminHandler:  adminHandler,
		SubHandler:    subHandler,
		CaseHandler:   caseHandler,
		RecHandler:    recHandler,
		TransHandler:  transHandler,
		ReportHandler: reportHandler,
	})

	// 8. Start Listening
	addr := ":" + cfg.Port
	log.Printf("[INFO] Server listening on %s (Environment: %s)\n", addr, cfg.Environment)
	if err := app.Listen(addr); err != nil {
		log.Fatalf("[FATAL] Server error: %v", err)
	}
}

