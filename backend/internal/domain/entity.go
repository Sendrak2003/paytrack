// Package domain содержит чистые бизнес-сущности и value objects.
// Нет зависимостей от фреймворков (GORM, Gin и т.д.).
package domain

import "time"

// Client — юридическое лицо / контрагент.
type Client struct {
	ID            uint
	Name          string
	INN           string
	OGRN          string
	BankAccount   string
	ContactPerson string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// Project — проект или направление работ.
type Project struct {
	ID       uint
	Name     string
	ClientID uint
	Client   *Client
	Status   string // active, completed, paused
}

// Payment — входящая оплата.
type Payment struct {
	ID             uint
	ProjectID      uint
	Project        *Project
	LegalEntityID  uint
	LegalEntity    *Client
	PaymentDate    time.Time
	Amount         float64
	PaymentPurpose string
	ServiceStage   string
	InvoiceNumber  string
	ContractNumber string
	ExternalID     string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	Act            *Act
}

// Act — закрывающий документ (акт выполненных работ).
type Act struct {
	ID             uint
	PaymentID      uint
	IsSent         bool
	SentAt         *time.Time
	IsSigned       bool
	SignedAt       *time.Time
	Status         ActStatus
	ManagerComment string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// ActStatus — вычисляемый статус акта.
type ActStatus string

const (
	ActStatusNotSent          ActStatus = "not_sent"
	ActStatusWaitingSignature ActStatus = "waiting_signature"
	ActStatusClosed           ActStatus = "closed"
	ActStatusNeedsAttention   ActStatus = "needs_attention"
)

// CalculateStatus — бизнес-логика расчёта статуса акта.
// Это domain logic, не зависит от инфраструктуры.
func (a *Act) CalculateStatus(paymentDate time.Time) ActStatus {
	if a.IsSent && a.IsSigned {
		return ActStatusClosed
	}
	if a.IsSent && !a.IsSigned {
		if a.SentAt != nil && time.Since(*a.SentAt) > 14*24*time.Hour {
			return ActStatusNeedsAttention
		}
		return ActStatusWaitingSignature
	}
	if time.Since(paymentDate) > 30*24*time.Hour {
		return ActStatusNeedsAttention
	}
	return ActStatusNotSent
}

// DashboardSummary — агрегат для дашборда.
type DashboardSummary struct {
	TotalPayments         float64
	TotalPaymentsCount    int
	TotalProjects         int
	ClosedActsAmount      float64
	OpenActsAmount        float64
	NotSentCount          int
	WaitingSignatureCount int
	NeedsAttentionCount   int
}

// ProjectSummary — агрегат для списка проектов.
type ProjectSummary struct {
	ID              uint
	Name            string
	ClientName      string
	INN             string
	Status          string
	TotalAmount     float64
	PaymentsCount   int
	ClosedActsCount int
	OpenActsCount   int
	DocStatus       string // all_closed, has_open, needs_attention
}

// PaymentFilter — критерии фильтрации оплат.
type PaymentFilter struct {
	ProjectID     *uint
	LegalEntityID *uint
	ActStatus     *string
	ServiceStage  *string
	Search        *string
	DateFrom      *string
	DateTo        *string
}
