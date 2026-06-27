package domain

// Repository — порт (интерфейс) для доступа к данным.
// Реализация в infrastructure/persistence — inject через DI.
type Repository interface {
	// Clients
	GetAllClients() ([]Client, error)

	// Projects
	GetProjectSummaries() ([]ProjectSummary, error)
	GetProjectByID(id uint) (*Project, error)

	// Payments
	GetPayments(filter PaymentFilter) ([]Payment, error)
	GetPaymentByID(id uint) (*Payment, error)

	// Acts
	UpsertAct(paymentID uint, isSent, isSigned bool, comment string) (*Act, error)

	// Dashboard
	GetDashboardSummary() (*DashboardSummary, error)
}
