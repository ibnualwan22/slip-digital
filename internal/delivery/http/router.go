package http

import (
	"github.com/ibnualwan/bisyaroh/pkg/response"
	"github.com/labstack/echo/v4"
)

func SetupRouter(
	e *echo.Echo,
	jwtSecret string,
	authHandler *AuthHandler,
	employeeHandler *EmployeeHandler,
	activityHandler *ActivityHandler,
	payrollHandler *PayrollHandler,
	categoryHandler *CategoryHandler,
	siakadHandler *SiakadHandler,
	expenseHandler *ExpenseHandler,
	dashboardHandler *DashboardHandler,
	periodHandler *PeriodHandler,
	meHandler *MeHandler,
	kbmHandler *KBMHandler,
	ssoHandler *SSOHandler,
	ssoBawwabahHandler *SSOHandler,
) {
	// Health check / Root endpoint
	e.GET("/", func(c echo.Context) error {
		return response.Success(c, 200, "Bisyaroh API is running", nil)
	})

	api := e.Group("/api/v1")

	// Protected routes
	// Login bersifat publik; semua rute lain di bawah grup auth wajib JWT.
	api.POST("/auth/login", authHandler.Login)

	// Webhook dari Siakad (publik, dijaga secret bersama di handler).
	if kbmHandler != nil {
		api.POST("/webhooks/siakad", kbmHandler.Webhook)
	}

	// Callback SSO dari Siakad (publik; token sekali pakai diverifikasi di handler).
	if ssoHandler != nil {
		api.GET("/sso/siakad", ssoHandler.Callback)
	}

	// Callback SSO dari Bawwabah (publik; token sekali pakai diverifikasi di handler).
	if ssoBawwabahHandler != nil {
		api.GET("/sso/bawwabah", ssoBawwabahHandler.Callback)
	}

	auth := api.Group("")
	auth.Use(JWTMiddleware(jwtSecret))
	auth.Use(RequireRole("admin"))

	// Pemetaan Periode Dufah (khusus admin; grup auth sudah berpagar peran admin)
	if periodHandler != nil {
		periods := auth.Group("/periods")
		periods.GET("", periodHandler.List)
		periods.POST("", periodHandler.Create)
		periods.PUT("/:id", periodHandler.Update)
		periods.DELETE("/:id", periodHandler.Delete)
	}

	// Rekap Saya (khusus pengajar) — grup terpisah, pagar peran pengajar.
	// Identitas selalu dari token; tidak ada parameter identitas dari klien.
	if meHandler != nil {
		me := api.Group("/me")
		me.Use(JWTMiddleware(jwtSecret))
		me.Use(RequireRole("pengajar"))
		me.GET("/profile", meHandler.Profile)
		me.GET("/periods", meHandler.Periods)
		me.GET("/recap", meHandler.Recap)
	}

	// Dashboard
	if dashboardHandler != nil {
		dashboard := auth.Group("/dashboard")
		dashboard.GET("/stats", dashboardHandler.GetStats)
	}

	// Employees
	employees := auth.Group("/employees")
	employees.GET("", employeeHandler.List)
	employees.POST("", employeeHandler.Create)
	employees.GET("/:id", employeeHandler.Get)
	employees.PUT("/:id", employeeHandler.Update)
	employees.DELETE("/:id", employeeHandler.Delete)

	// Activities
	activities := auth.Group("/activities")
	activities.GET("", activityHandler.List)
	activities.POST("", activityHandler.Create)
	activities.GET("/:id", activityHandler.Get)
	activities.PUT("/:id", activityHandler.Update)

	// Payroll
	payroll := auth.Group("/payroll")
	if kbmHandler != nil {
		payroll.POST("/recompute", kbmHandler.Recompute)
		payroll.GET("/kbm-info", kbmHandler.Info)
	}
	payroll.GET("", payrollHandler.ListTransactions)
	payroll.POST("", payrollHandler.CreateTransaction)
	payroll.POST("/generate", payrollHandler.GeneratePayroll)
	payroll.POST("/sync-employees", payrollHandler.SyncEmployees)
	payroll.POST("/bulk-send", payrollHandler.BulkSendWA)
	payroll.GET("/:id", payrollHandler.GetTransaction)
	payroll.DELETE("/:id", payrollHandler.DeleteTransaction)

	payroll.GET("/:id/details", payrollHandler.ListDetails)
	payroll.POST("/:id/details", payrollHandler.AddDetail)
	payroll.PUT("/details/:detailId", payrollHandler.UpdateDetail)
	payroll.DELETE("/details/:detailId", payrollHandler.RemoveDetail)
	payroll.POST("/:id/calculate", payrollHandler.CalculateTHP)
	payroll.POST("/:id/status", payrollHandler.UpdateStatus)
	payroll.GET("/:id/wa-preview", payrollHandler.PreviewSlipWA)
	payroll.POST("/:id/send-wa", payrollHandler.SendSlipWA)

	// Category Routes
	categories := auth.Group("/categories")
	categories.GET("", categoryHandler.List)
	categories.POST("", categoryHandler.Create)
	categories.GET("/:id", categoryHandler.Get)
	categories.PUT("/:id", categoryHandler.Update)
	categories.DELETE("/:id", categoryHandler.Delete)

	// Siakad Routes
	if siakadHandler != nil {
		siakad := auth.Group("/siakad")
		siakad.GET("/pengajar", siakadHandler.GetPengajar)
		siakad.GET("/dufah", siakadHandler.GetDufah)
		siakad.PUT("/pengajar/terlambat", siakadHandler.UpdateTerlambat)
		siakad.POST("/sync-all", siakadHandler.SyncAllToPayroll)
		siakad.POST("/pengajar/:siakadId/sync", siakadHandler.SyncToPayroll)
	}

	// Expense (Laporan Pengeluaran) Routes
	if expenseHandler != nil {
		expenses := auth.Group("/expenses")
		expenses.GET("", expenseHandler.ListReports)
		expenses.POST("", expenseHandler.CreateReport)
		expenses.GET("/:id", expenseHandler.GetReport)
		expenses.DELETE("/:id", expenseHandler.DeleteReport)
		expenses.POST("/:id/items", expenseHandler.AddItem)
		expenses.PUT("/items/:itemId", expenseHandler.UpdateItem)
		expenses.DELETE("/items/:itemId", expenseHandler.DeleteItem)
		expenses.POST("/:id/tashih-all", expenseHandler.TashihAll)
		expenses.POST("/upload", expenseHandler.UploadBukti)
		expenses.POST("/scan-receipt", expenseHandler.ScanReceipt)
		expenses.POST("/scan-receipt-url", expenseHandler.ScanReceiptFromURL)
	}
}
