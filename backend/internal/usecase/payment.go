package usecase

import "payments-dashboard/internal/domain"

// PaymentUseCase — use case для работы с оплатами и актами.
type PaymentUseCase struct {
	repo domain.Repository
}

func NewPaymentUseCase(repo domain.Repository) *PaymentUseCase {
	return &PaymentUseCase{repo: repo}
}

func (uc *PaymentUseCase) GetPayments(filter domain.PaymentFilter) ([]domain.Payment, error) {
	return uc.repo.GetPayments(filter)
}

func (uc *PaymentUseCase) GetPaymentByID(id uint) (*domain.Payment, error) {
	return uc.repo.GetPaymentByID(id)
}

func (uc *PaymentUseCase) UpdateAct(paymentID uint, isSent, isSigned bool, comment string) (*domain.Act, error) {
	return uc.repo.UpsertAct(paymentID, isSent, isSigned, comment)
}
