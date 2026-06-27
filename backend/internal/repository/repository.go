package repository

import (
	"errors"
	"fmt"
	"log/slog"
	"payments-dashboard/internal/models"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	maxLockRetries = 5
	lockRetryDelay = 50 * time.Millisecond
)

type Repository struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// isLockConflict reports whether an error is a transient lock/serialization
// failure worth retrying. Covers SQLite ("database is locked", "table is
// locked", SQLITE_BUSY) and Postgres serialization failures (SQLSTATE 40001/40P01).
func isLockConflict(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "database is locked"),
		strings.Contains(msg, "table is locked"),
		strings.Contains(msg, "sqlite_busy"),
		strings.Contains(msg, "deadlock detected"),
		strings.Contains(msg, "could not serialize"),
		strings.Contains(msg, "40001"),
		strings.Contains(msg, "40p01"):
		return true
	}
	return false
}

// retryOnLockConflict runs fn, retrying with a small backoff while it fails
// with a transient lock conflict. Non-lock errors are returned immediately.
func retryOnLockConflict(attempts int, fn func() error) error {
	var err error
	for i := 0; i < attempts; i++ {
		err = fn()
		if err == nil || !isLockConflict(err) {
			return err
		}
		slog.Warn("lock conflict, retrying transaction",
			"attempt", i+1, "max", attempts, "error", err)
		time.Sleep(lockRetryDelay * time.Duration(i+1)) // linear backoff
	}
	return err
}

// ---- Clients ----

func (r *Repository) GetAllClients() ([]models.Client, error) {
	var clients []models.Client
	err := r.db.Find(&clients).Error
	return clients, err
}

// ---- Projects ----

func (r *Repository) GetProjectSummaries() ([]models.ProjectSummary, error) {
	var projects []models.Project
	err := r.db.Preload("Client").Preload("Payments.Act").Find(&projects).Error
	if err != nil {
		return nil, err
	}

	summaries := make([]models.ProjectSummary, 0, len(projects))
	for _, p := range projects {
		s := models.ProjectSummary{
			ID:     p.ID,
			Name:   p.Name,
			Status: p.Status,
		}
		if p.Client != nil {
			s.ClientName = p.Client.Name
			s.INN = p.Client.INN
		}
		s.PaymentsCount = len(p.Payments)
		hasNeedsAttention := false
		for _, pay := range p.Payments {
			s.TotalAmount += pay.Amount
			if pay.Act != nil {
				st := pay.Act.CalculateStatus(pay.PaymentDate)
				if st == "closed" {
					s.ClosedActsCount++
				} else {
					s.OpenActsCount++
					if st == "needs_attention" {
						hasNeedsAttention = true
					}
				}
			} else {
				s.OpenActsCount++
				if time.Since(pay.PaymentDate) > 30*24*time.Hour {
					hasNeedsAttention = true
				}
			}
		}
		if s.OpenActsCount == 0 && s.PaymentsCount > 0 {
			s.DocStatus = "all_closed"
		} else if hasNeedsAttention {
			s.DocStatus = "needs_attention"
		} else {
			s.DocStatus = "has_open"
		}
		summaries = append(summaries, s)
	}
	return summaries, nil
}

func (r *Repository) GetProjectByID(id uint) (*models.Project, error) {
	var project models.Project
	err := r.db.Preload("Client").First(&project, id).Error
	return &project, err
}

// ---- Payments ----

func (r *Repository) GetPayments(f models.PaymentFilter) ([]models.Payment, error) {
	var payments []models.Payment
	q := r.db.Preload("Project.Client").Preload("LegalEntity").Preload("Act")

	if f.ProjectID != nil {
		q = q.Where("payments.project_id = ?", *f.ProjectID)
	}
	if f.LegalEntityID != nil {
		q = q.Where("payments.legal_entity_id = ?", *f.LegalEntityID)
	}
	if f.ServiceStage != nil && *f.ServiceStage != "" {
		q = q.Where("payments.service_stage = ?", *f.ServiceStage)
	}
	if f.DateFrom != nil && *f.DateFrom != "" {
		q = q.Where("payments.payment_date >= ?", *f.DateFrom)
	}
	if f.DateTo != nil && *f.DateTo != "" {
		q = q.Where("payments.payment_date <= ?", *f.DateTo)
	}
	if f.Search != nil && *f.Search != "" {
		like := fmt.Sprintf("%%%s%%", *f.Search)
		q = q.Joins("LEFT JOIN clients ON clients.id = payments.legal_entity_id").
			Where("payments.payment_purpose LIKE ? OR clients.name LIKE ?", like, like)
	}

	err := q.Order("payments.payment_date DESC").Find(&payments).Error
	if err != nil {
		return nil, err
	}

	// Filter by act status (in-memory because status is computed)
	if f.ActStatus != nil && *f.ActStatus != "" {
		filtered := payments[:0]
		for _, pay := range payments {
			var st string
			if pay.Act != nil {
				st = pay.Act.CalculateStatus(pay.PaymentDate)
				// Update computed status
				pay.Act.Status = st
			} else {
				if time.Since(pay.PaymentDate) > 30*24*time.Hour {
					st = "needs_attention"
				} else {
					st = "not_sent"
				}
			}
			if st == *f.ActStatus {
				filtered = append(filtered, pay)
			}
		}
		payments = filtered
	} else {
		// Always compute status
		for i := range payments {
			if payments[i].Act != nil {
				payments[i].Act.Status = payments[i].Act.CalculateStatus(payments[i].PaymentDate)
			}
		}
	}

	return payments, nil
}

func (r *Repository) GetPaymentByID(id uint) (*models.Payment, error) {
	var payment models.Payment
	err := r.db.Preload("Project.Client").Preload("LegalEntity").Preload("Act").First(&payment, id).Error
	return &payment, err
}

// ---- Acts ----

func (r *Repository) GetActByPaymentID(paymentID uint) (*models.Act, error) {
	var act models.Act
	err := r.db.Where("payment_id = ?", paymentID).First(&act).Error
	if err != nil {
		return nil, err
	}
	return &act, nil
}

// UpsertAct creates or updates the act for a payment.
//
// Concurrency: two managers may hit "акт подписан" at the same time for the
// same payment. To avoid a lost update / race condition we:
//  1. wrap the read-modify-write in a single transaction;
//  2. lock the row with SELECT ... FOR UPDATE (clause.Locking). On SQLite this
//     degrades to a serialized write transaction, which is still correct;
//  3. retry the whole transaction on a transient lock conflict
//     ("database is locked" / "SQLITE_BUSY", or a Postgres serialization error).
//
// This is the one place in the project that genuinely needs row locking —
// every other operation is a single-statement read or write.
func (r *Repository) UpsertAct(paymentID uint, isSent, isSigned bool, comment string) (*models.Act, error) {
	var result *models.Act

	err := retryOnLockConflict(maxLockRetries, func() error {
		return r.db.Transaction(func(tx *gorm.DB) error {
			now := time.Now()

			// Lock the payment row first to establish a stable ordering and
			// guarantee the payment exists. clause.Locking is a no-op on SQLite
			// but produces FOR UPDATE on Postgres/MySQL.
			var pay models.Payment
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
				First(&pay, paymentID).Error; err != nil {
				return err
			}

			// Lock the existing act row (if any).
			var act models.Act
			findErr := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
				Where("payment_id = ?", paymentID).First(&act).Error

			isNew := errors.Is(findErr, gorm.ErrRecordNotFound)
			if findErr != nil && !isNew {
				return findErr
			}

			if isNew {
				act = models.Act{PaymentID: paymentID}
			}

			// Apply changes. Set timestamps only on a real transition.
			if isSent && !act.IsSent {
				act.SentAt = &now
			}
			if !isSent {
				act.SentAt = nil
			}
			if isSigned && !act.IsSigned {
				act.SignedAt = &now
			}
			if !isSigned {
				act.SignedAt = nil
			}
			act.IsSent = isSent
			act.IsSigned = isSigned
			act.ManagerComment = comment
			act.Status = act.CalculateStatus(pay.PaymentDate)

			if isNew {
				if err := tx.Create(&act).Error; err != nil {
					return err
				}
			} else {
				if err := tx.Save(&act).Error; err != nil {
					return err
				}
			}

			result = &act
			return nil
		})
	})

	return result, err
}

// ---- Dashboard summary ----

func (r *Repository) GetDashboardSummary() (*models.DashboardSummary, error) {
	var payments []models.Payment
	err := r.db.Preload("Act").Find(&payments).Error
	if err != nil {
		return nil, err
	}

	s := &models.DashboardSummary{}

	projectIDs := map[uint]bool{}
	for _, pay := range payments {
		s.TotalPayments += pay.Amount
		s.TotalPaymentsCount++
		projectIDs[pay.ProjectID] = true

		var st string
		if pay.Act != nil {
			st = pay.Act.CalculateStatus(pay.PaymentDate)
		} else {
			if time.Since(pay.PaymentDate) > 30*24*time.Hour {
				st = "needs_attention"
			} else {
				st = "not_sent"
			}
		}

		switch st {
		case "closed":
			s.ClosedActsAmount += pay.Amount
		case "not_sent":
			s.OpenActsAmount += pay.Amount
			s.NotSentCount++
		case "waiting_signature":
			s.OpenActsAmount += pay.Amount
			s.WaitingSignatureCount++
		case "needs_attention":
			s.OpenActsAmount += pay.Amount
			s.NeedsAttentionCount++
		}
	}
	s.TotalProjects = len(projectIDs)
	return s, nil
}
