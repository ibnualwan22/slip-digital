package main

import (
	"log"

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
	e.Use(middleware.CORS())

	// Initialize Repositories
	empRepo := postgres.NewEmployeeRepository(cfg.DB)
	actRepo := postgres.NewActivityRepository(cfg.DB)
	payRepo := postgres.NewPayrollRepository(cfg.DB)
	catRepo := postgres.NewCategoryRepository(cfg.DB)

	// Note: We need to initialize payrollService first to inject it into employeeService but 
	// payrollService also needs empRepo.
	payService := usecase.NewPayrollService(payRepo, empRepo)
	empService := usecase.NewEmployeeService(empRepo, payService)

	// Initialize WA Services
	waClient := whatsapp.NewWhatsAppClient(cfg.WA_API_URL, cfg.WA_API_KEY, cfg.WA_SESSION)
	slipGen, err := generator.NewSlipGenerator("web/assets/img/sampul-slip.png", "web/assets/img/stempel-slip.png")
	if err != nil {
		log.Fatalf("Failed to initialize slip generator: %v", err)
	}

	// Initialize Handlers
	empHandler := delivery.NewEmployeeHandler(empService)
	actHandler := delivery.NewActivityHandler(actRepo)
	payHandler := delivery.NewPayrollHandler(payService, waClient, slipGen)
	catHandler := delivery.NewCategoryHandler(catRepo)

	// Setup Routes
	delivery.SetupRouter(e, cfg.JWTSecret, empHandler, actHandler, payHandler, catHandler)

	// Serve Frontend (static files)
	e.Static("/assets", "web/assets")
	e.File("/", "web/index.html")
	// e.File("/*", "web/index.html") // Catch all for SPA if needed, but might conflict with unhandled API routes if not careful. For hash routing, we don't strictly need it, but good to have.
	// Actually, with hash routing (#/dashboard), we only need / to serve index.html

	// Start Server
	log.Printf("Starting server on port %s", cfg.Port)
	if err := e.Start(":" + cfg.Port); err != nil {
		log.Fatalf("Server stopped: %v", err)
	}
}
