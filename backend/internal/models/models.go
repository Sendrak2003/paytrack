package models

import (
	"time"
)

type Client struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	Name          string    `json:"name" gorm:"not null"`
	INN           string    `json:"inn"`
	OGRN          string    `json:"ogrn"`
	BankAccount   string    `json:"bank_account"`
	ContactPerson string    `json:"contact_person"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	Projects      []Project `json:"projects,omitempty" gorm:"foreignKey:ClientID"`
}

type Project struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name" gorm:"not null"`
	ClientID  uint      `json:"client_id"`
	Client    *Client   `json:"client,omitempty" gorm:"foreignKey:ClientID"`
	Status    string    `json:"status" gorm:"default:'active'"` // active, completed, paused
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Payments  []Payment `json:"payments,omitempty" gorm:"foreignKey:ProjectID"`
}

type Payment struct {
	ID             uint      `json:"id" gorm:"primaryKey"`
	ProjectID      uint      `json:"project_id"`
	Project        *Project  `json:"project,omitempty" gorm:"foreignKey:ProjectID"`
	LegalEntityID  uint      `json:"legal_entity_id"`
	LegalEntity    *Client   `json:"legal_entity,omitempty" gorm:"foreignKey:LegalEntityID"`
	PaymentDate    time.Time `json:"payment_date"`
	Amount         float64   `json:"amount" gorm:"not null"`
	PaymentPurpose string    `json:"payment_purpose"`
	ServiceStage   string    `json:"service_stage"` // разработка, SEO, реклама, дизайн, контент, сопровождение
	InvoiceNumber  string    `json:"invoice_number"`
	ContractNumber string    `json:"contract_number"`
	// ExternalID is a stable, unique key derived from the source bank operation
	// (date + amount + payer INN + purpose). It makes import idempotent:
	// re-importing the same statement never duplicates a payment.
	ExternalID string    `json:"external_id" gorm:"uniqueIndex"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	Act        *Act      `json:"act,omitempty" gorm:"foreignKey:PaymentID"`
}

type Act struct {
	ID             uint       `json:"id" gorm:"primaryKey"`
	PaymentID      uint       `json:"payment_id" gorm:"uniqueIndex"`
	Payment        *Payment   `json:"payment,omitempty" gorm:"foreignKey:PaymentID"`
	IsSent         bool       `json:"is_sent" gorm:"default:false"`
	SentAt         *time.Time `json:"sent_at"`
	IsSigned       bool       `json:"is_signed" gorm:"default:false"`
	SignedAt        *time.Time `json:"signed_at"`
	Status         string     `json:"status"` // not_sent, waiting_signature, closed, needs_attention
	ManagerComment string     `json:"manager_comment"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// Status calculation logic
func (a *Act) CalculateStatus(paymentDate time.Time) string {
	if a.IsSent && a.IsSigned {
		return "closed"
	}
	if a.IsSent && !a.IsSigned {
		// If sent more than 14 days ago without signature — needs attention
		if a.SentAt != nil && time.Since(*a.SentAt) > 14*24*time.Hour {
			return "needs_attention"
		}
		return "waiting_signature"
	}
	// Not sent — check if payment is old (>30 days)
	if time.Since(paymentDate) > 30*24*time.Hour {
		return "needs_attention"
	}
	return "not_sent"
}

// Summary DTO
type DashboardSummary struct {
	TotalPayments        float64 `json:"total_payments"`
	TotalPaymentsCount   int     `json:"total_payments_count"`
	TotalProjects        int     `json:"total_projects"`
	ClosedActsAmount     float64 `json:"closed_acts_amount"`
	OpenActsAmount       float64 `json:"open_acts_amount"`
	NotSentCount         int     `json:"not_sent_count"`
	WaitingSignatureCount int    `json:"waiting_signature_count"`
	NeedsAttentionCount  int     `json:"needs_attention_count"`
}

// Project summary DTO
type ProjectSummary struct {
	ID              uint    `json:"id"`
	Name            string  `json:"name"`
	ClientName      string  `json:"client_name"`
	INN             string  `json:"inn"`
	Status          string  `json:"status"`
	TotalAmount     float64 `json:"total_amount"`
	PaymentsCount   int     `json:"payments_count"`
	ClosedActsCount int     `json:"closed_acts_count"`
	OpenActsCount   int     `json:"open_acts_count"`
	DocStatus       string  `json:"doc_status"` // all_closed, has_open, needs_attention
}

// Payment with act filter params
type PaymentFilter struct {
	ProjectID     *uint   `form:"project_id"`
	LegalEntityID *uint   `form:"legal_entity_id"`
	ActStatus     *string `form:"act_status"`
	ServiceStage  *string `form:"service_stage"`
	Search        *string `form:"search"`
	DateFrom      *string `form:"date_from"`
	DateTo        *string `form:"date_to"`
}
