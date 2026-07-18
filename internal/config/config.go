package config

import (
	"fmt"
	"log"
	"os"

	"github.com/google/uuid"
	"github.com/ibnualwan/bisyaroh/internal/domain"
	"github.com/joho/godotenv"
	"github.com/shopspring/decimal"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Config struct {
	Port       string
	JWTSecret  string
	DB         *gorm.DB
	WA_API_URL string
	WA_API_KEY string
	WA_SESSION string

	SiakadAPIURL string
	SiakadAPIKey string

	// Gemini AI
	GeminiAPIKey string

	// Cloudinary
	CloudinaryCloudName string
	CloudinaryAPIKey    string
	CloudinaryAPISecret string
}

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, relying on environment variables")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "defaultsecret"
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_SSLMODE"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Migrate and Seed
	migrateAndSeed(db)

	return &Config{
		Port:       port,
		JWTSecret:  jwtSecret,
		DB:         db,
		WA_API_URL: os.Getenv("WA_API_URL"),
		WA_API_KEY: os.Getenv("WA_API_KEY"),
		WA_SESSION: os.Getenv("WA_SESSION_ID"),
		SiakadAPIURL: os.Getenv("SIAKAD_API_URL"),
		SiakadAPIKey: os.Getenv("SIAKAD_API_KEY"),
		GeminiAPIKey:        os.Getenv("API_GEMINI"),
		CloudinaryCloudName: os.Getenv("CLOUDE_NAME"),
		CloudinaryAPIKey:    os.Getenv("API_KEY_CLOUDINARY"),
		CloudinaryAPISecret: os.Getenv("API_SECRET_CLOUDINARY"),
	}
}

func migrateAndSeed(db *gorm.DB) {
	err := db.AutoMigrate(
		&domain.EmployeeCategory{},
		&domain.MasterActivity{},
		&domain.Employee{},
		&domain.PayrollTransaction{},
		&domain.PayrollDetail{},
		&domain.ExpenseReport{},
		&domain.ExpenseItem{},
	)
	if err != nil {
		log.Printf("Failed to auto migrate database schema: %v\n", err)
	}

	// Seed Data
	categories := []domain.EmployeeCategory{
		{
			ID:                  uuid.MustParse("00000000-0000-0000-0000-000000000001"),
			Code:                "BUK",
			Name:                "Biro Urusan Kepesantrenan (BUK)",
			CalcMethod:          "HOURLY",
			FixedSalary:         decimal.Zero,
			StructuralAllowance: decimal.Zero,
			TargetIncentive:     decimal.Zero,
			HourlyRate:          decimal.Zero, // To be configured by user
		},
		{
			ID:                  uuid.MustParse("00000000-0000-0000-0000-000000000002"),
			Code:                "S2_GEL1",
			Name:                "S2 Gelombang 1",
			CalcMethod:          "PROPORTIONAL",
			FixedSalary:         decimal.Zero,
			StructuralAllowance: decimal.Zero,
			TargetIncentive:     decimal.Zero, // To be configured by user
			HourlyRate:          decimal.Zero,
		},
		{
			ID:                  uuid.MustParse("00000000-0000-0000-0000-000000000003"),
			Code:                "S2_GEL2",
			Name:                "S2 Gelombang 2",
			CalcMethod:          "PROPORTIONAL",
			FixedSalary:         decimal.Zero,
			StructuralAllowance: decimal.Zero,
			TargetIncentive:     decimal.Zero,
			HourlyRate:          decimal.Zero,
		},
		{
			ID:                  uuid.MustParse("00000000-0000-0000-0000-000000000004"),
			Code:                "REGULER",
			Name:                "Reguler",
			CalcMethod:          "FIXED",
			FixedSalary:         decimal.Zero,
			StructuralAllowance: decimal.Zero,
			TargetIncentive:     decimal.Zero,
			HourlyRate:          decimal.Zero,
		},
		{
			ID:                  uuid.MustParse("00000000-0000-0000-0000-000000000005"),
			Code:                "KSU",
			Name:                "Kasir & Keuangan (KSU)",
			CalcMethod:          "HOURLY",
			FixedSalary:         decimal.Zero,
			StructuralAllowance: decimal.Zero,
			TargetIncentive:     decimal.Zero,
			HourlyRate:          decimal.Zero,
		},
	}

	for _, cat := range categories {
		var existing domain.EmployeeCategory
		if err := db.Where("code = ?", cat.Code).First(&existing).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				log.Printf("Seeding category: %s", cat.Code)
				db.Create(&cat)
			}
		}
	}
}
