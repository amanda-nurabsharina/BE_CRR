package http

import (
	"be_crr/internal/adapter/inbound/http/handler"
	"be_crr/internal/adapter/inbound/http/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

type RouterDependencies struct {
	JWTSecret       string
	AuthHandler     *handler.AuthHandler
	AdminHandler    *handler.AdminHandler
	SubHandler      *handler.SubscriptionHandler
	CaseHandler     *handler.CaseHandler
	RecHandler      *handler.RecordingHandler
	TransHandler    *handler.TranscriptHandler
	ReportHandler   *handler.ReportHandler
}

func SetupRouter(app *fiber.App, deps RouterDependencies) {
	// Global Middlewares
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "*",
		AllowMethods:     "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization, Range",
		ExposeHeaders:    "Content-Range, Accept-Ranges, Content-Length",
		AllowCredentials: false,
	}))

	// Health Check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok", "service": "Court Recording Report API"})
	})

	api := app.Group("/api/v1")

	// Public Routes
	auth := api.Group("/auth")
	auth.Post("/register", deps.AuthHandler.Register)
	auth.Post("/login", deps.AuthHandler.Login)

	public := api.Group("/public")
	public.Get("/verify/:token", deps.ReportHandler.PublicVerify)

	// Protected Routes (Requires JWT)
	authMW := middleware.AuthMiddleware(deps.JWTSecret)

	api.Get("/auth/me", authMW, deps.AuthHandler.GetMe)
	api.Put("/auth/profile", authMW, deps.AuthHandler.UpdateProfile)

	// Super Admin Routes
	admin := api.Group("/admin", authMW, middleware.SuperAdminOnly())
	admin.Get("/metrics", deps.AdminHandler.GetMetrics)
	admin.Get("/tenants", deps.AdminHandler.ListTenants)
	admin.Patch("/tenants/:id/status", deps.AdminHandler.UpdateTenantStatus)
	admin.Post("/tenants/:id/quota", deps.AdminHandler.AdjustQuota)
	admin.Get("/payments/pending", deps.AdminHandler.ListPendingPayments)
	admin.Post("/payments/:id/approve", deps.AdminHandler.ApprovePayment)
	admin.Post("/payments/:id/reject", deps.AdminHandler.RejectPayment)

	// Tenant User Routes
	tenant := api.Group("", authMW, middleware.TenantUserOnly())

	// Subscriptions
	tenant.Get("/subscriptions/current", deps.SubHandler.GetCurrent)
	tenant.Post("/subscriptions/manual-payment", deps.SubHandler.SubmitManualPayment)
	tenant.Get("/subscriptions/history", deps.SubHandler.GetPaymentHistory)

	// Court Cases & Sessions
	tenant.Get("/cases", deps.CaseHandler.ListCases)
	tenant.Post("/cases", deps.CaseHandler.CreateCase)
	tenant.Get("/cases/:id", deps.CaseHandler.GetCase)
	tenant.Delete("/cases/:id/purge", deps.CaseHandler.ManualPurge)
	tenant.Post("/cases/:id/sessions", deps.CaseHandler.CreateSession)
	tenant.Get("/cases/:id/sessions/:sessionId", deps.CaseHandler.GetSession)

	// Audio Ingest & Recording
	tenant.Post("/recording/sessions", deps.RecHandler.InitSession)
	tenant.Post("/recording/chunk", deps.RecHandler.UploadChunk)
	tenant.Post("/recording/sessions/:id/finalize", deps.RecHandler.FinalizeSession)
	tenant.Post("/ingest/upload", deps.RecHandler.UploadMediaFile)
	tenant.Get("/evidence/:id/stream", deps.RecHandler.StreamAudio)

	// Transcripts
	tenant.Post("/evidence/:audioId/transcribe", deps.TransHandler.TriggerASR)
	tenant.Get("/evidence/:audioId/transcripts", deps.TransHandler.GetTranscript)
	tenant.Put("/transcripts/segments", deps.TransHandler.SaveSegments)

	// Reports (Berita Acara Sidang)
	tenant.Post("/reports", deps.ReportHandler.CreateReport)
	tenant.Get("/reports/:id", deps.ReportHandler.GetReport)
	tenant.Put("/reports/:id", deps.ReportHandler.UpdateReport)
	tenant.Post("/reports/:id/sign", deps.ReportHandler.SignReport)
}

