package http

import (
	"github.com/ibnualwan/bisyaroh/pkg/response"
	"github.com/labstack/echo/v4"
)

func SetupRouter(
	e *echo.Echo,
	jwtSecret string,
	employeeHandler *EmployeeHandler,
	activityHandler *ActivityHandler,
	payrollHandler *PayrollHandler,
	categoryHandler *CategoryHandler,
	siakadHandler *SiakadHandler,
	expenseHandler *ExpenseHandler,
) {
	// Health check / Root endpoint
	e.GET("/", func(c echo.Context) error {
		return response.Success(c, 200, "Bisyaroh API is running", nil)
	})

	api := e.Group("/api/v1")

	// Protected routes
	auth := api.Group("")
	// We comment out the JWT middleware for development/testing convenience.
	// auth.Use(JWTMiddleware(jwtSecret))

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
	payroll.GET("", payrollHandler.ListTransactions)
	payroll.POST("", payrollHandler.CreateTransaction)
	payroll.POST("/generate", payrollHandler.GeneratePayroll)
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
	categories := api.Group("/categories")
	categories.GET("", categoryHandler.List)
	categories.POST("", categoryHandler.Create)
	categories.GET("/:id", categoryHandler.Get)
	categories.PUT("/:id", categoryHandler.Update)
	categories.DELETE("/:id", categoryHandler.Delete)

	// Siakad Routes
	if siakadHandler != nil {
		siakad := api.Group("/siakad")
		siakad.GET("/pengajar", siakadHandler.GetPengajar)
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
		expenses.DELETE("/items/:itemId", expenseHandler.DeleteItem)
		expenses.POST("/:id/tashih-all", expenseHandler.TashihAll)
		expenses.POST("/upload", expenseHandler.UploadBukti)
		expenses.POST("/scan-receipt", expenseHandler.ScanReceipt)
		expenses.POST("/scan-receipt-url", expenseHandler.ScanReceiptFromURL)
	}
}
