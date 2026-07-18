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

	// Setup Routes
	delivery.SetupRouter(e, cfg.JWTSecret, empHandler, actHandler, payHandler, catHandler, siakadHandler)

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
