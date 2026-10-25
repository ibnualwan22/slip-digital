package main

import (
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/ibnualwan/bisyaroh/internal/config"
	delivery "github.com/ibnualwan/bisyaroh/internal/delivery/http"
	"github.com/ibnualwan/bisyaroh/internal/repository/postgres"
	"github.com/ibnualwan/bisyaroh/internal/usecase"
	"github.com/ibnualwan/bisyaroh/pkg/generator"
	"github.com/ibnualwan/bisyaroh/pkg/whatsapp"
)

func main() {
	// Load Configuration
	cfg := config.LoadConfig()

	// Initialize Echo
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"https://e-maliyah.markazarabiyah.site"},
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
	}))

	// Initialize Repositories
	empRepo := postgres.NewEmployeeRepository(cfg.DB)
	periodRepo := postgres.NewPeriodRepository(cfg.DB)
	kbmLogRepo := postgres.NewKBMSyncLogRepository(cfg.DB)
	actRepo := postgres.NewActivityRepository(cfg.DB)
	payRepo := postgres.NewPayrollRepository(cfg.DB)
	catRepo := postgres.NewCategoryRepository(cfg.DB)
	expRepo := postgres.NewExpenseRepository(cfg.DB)
	ssoRepo := postgres.NewSSORepository(cfg.DB)

	// Note: We need to initialize payrollService first to inject it into employeeService but 
	// payrollService also needs empRepo.
	payService := usecase.NewPayrollService(payRepo, empRepo)
	empService := usecase.NewEmployeeService(empRepo, payService)

	// Initialize WA Services
	waClient := whatsapp.NewWhatsAppClient(cfg.WA_API_URL, cfg.WA_API_KEY, cfg.WA_SESSION)
	slipGen, err := generator.NewSlipGenerator("web/public/img/sampul-slip.png", "web/public/img/stempel-slip.png")
	if err != nil {
		log.Fatalf("Failed to initialize slip generator: %v", err)
	}

	// Initialize Handlers
	empHandler := delivery.NewEmployeeHandler(empService)
	actHandler := delivery.NewActivityHandler(actRepo)
	payHandler := delivery.NewPayrollHandler(payService, empRepo, waClient, slipGen)
	catHandler := delivery.NewCategoryHandler(catRepo)
	siakadHandler := delivery.NewSiakadHandler(cfg, empRepo, actRepo, payService)
	expService := usecase.NewExpenseService(expRepo, cfg.GeminiAPIKey)
	expHandler := delivery.NewExpenseHandler(expService, cfg.CloudinaryCloudName, cfg.CloudinaryAPIKey, cfg.CloudinaryAPISecret)

	kbmService := usecase.NewKBMService(cfg, empRepo, periodRepo, payRepo, actRepo, kbmLogRepo, payService)
	kbmHandler := delivery.NewKBMHandler(kbmService, periodRepo, kbmLogRepo, cfg)
	kbmHandler.StartReconciliation()

	periodHandler := delivery.NewPeriodHandler(periodRepo)
	meHandler := delivery.NewMeHandler(periodRepo, payRepo, cfg)
	ssoHandler := delivery.NewSSOHandler(empRepo, ssoRepo, cfg.JWTSecret, cfg.SSOSecret, false, "Siakad")
	ssoBawwabahHandler := delivery.NewSSOHandler(empRepo, ssoRepo, cfg.JWTSecret, cfg.BawwabahSSOSecret, true, "Bawwabah")

	dashRepo := postgres.NewDashboardRepository(cfg.DB)
	dashService := usecase.NewDashboardService(dashRepo)
	dashHandler := delivery.NewDashboardHandler(dashService)

	// Setup Routes
	authHandler := delivery.NewAuthHandler(cfg, empRepo)
	siakadHandler.SetKBMService(kbmService)
	delivery.SetupRouter(e, cfg.JWTSecret, authHandler, empHandler, actHandler, payHandler, catHandler, siakadHandler, expHandler, dashHandler, periodHandler, meHandler, kbmHandler, ssoHandler, ssoBawwabahHandler)

	// Serve static files (React build output)
	e.Static("/img", "web/public/img")
	e.Static("/assets", "web/dist/assets")
	e.File("/", "web/dist/index.html")
	e.File("/index.html", "web/dist/index.html")

	// SPA Fallback Handle for react-router (so refresh works on /employees etc)
	e.GET("/*", func(c echo.Context) error {
		return c.File("web/dist/index.html")
	})

	// Start Server
	log.Printf("Starting server on port %s", cfg.Port)
	if err := e.Start(":" + cfg.Port); err != nil {
		log.Fatalf("Server stopped: %v", err)
	}
}
