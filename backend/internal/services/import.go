package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"payments-dashboard/internal/ai"
	"payments-dashboard/internal/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ImportService turns incoming bank-statement data into Payments.
//
// Two entry points:
//   - Import(ops):      already-parsed operations (e.g. from a CRM/manual form);
//   - ImportRawText(s): raw statement text -> AI/regex extraction -> import.
//
// The valuable write-path properties hold for both:
//   - idempotency: each row has a deterministic ExternalID, so re-running an
//     import never duplicates payments;
//   - retry: the batch write is wrapped in a transaction and retried on
//     transient lock conflicts.
type ImportService struct {
	db        *gorm.DB
	log       *slog.Logger
	extractor ai.Extractor
}

func NewImportService(db *gorm.DB, log *slog.Logger, extractor ai.Extractor) *ImportService {
	return &ImportService{db: db, log: log, extractor: extractor}
}

// BankOperation is one parsed line from a statement.
type BankOperation struct {
	Date          string  `json:"date"`            // "2024-10-05"
	Amount        float64 `json:"amount"`
	PayerINN      string  `json:"payer_inn"`
	PayerName     string  `json:"payer_name"`
	Purpose       string  `json:"purpose"`
	InvoiceNumber string  `json:"invoice_number"`
}

type ImportResult struct {
	Imported  int      `json:"imported"`
	Skipped   int      `json:"skipped"` // already present (idempotent)
	Unmatched int      `json:"unmatched"`
	Errors    []string `json:"errors"`
}

const (
	importMaxRetries = 5
	importRetryDelay = 50 * time.Millisecond
)

// externalID builds a deterministic key for an operation so the same bank line
// always maps to the same payment.
func externalID(op BankOperation) string {
	raw := fmt.Sprintf("%s|%.2f|%s|%s", op.Date, op.Amount, op.PayerINN, op.Purpose)
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:16])
}

func isLockConflict(err error) bool {
	if err == nil {
		return false
	}
	m := strings.ToLower(err.Error())
	return strings.Contains(m, "database is locked") ||
		strings.Contains(m, "table is locked") ||
		strings.Contains(m, "deadlock") ||
		strings.Contains(m, "could not serialize") ||
		strings.Contains(m, "40001")
}

// Import processes a batch of bank operations.
func (s *ImportService) Import(ops []BankOperation) (*ImportResult, error) {
	res := &ImportResult{}

	for _, op := range ops {
		extID := externalID(op)

		err := s.importOne(op, extID, res)
		if err != nil {
			s.log.Error("import operation failed",
				"purpose", op.Purpose, "amount", op.Amount, "error", err)
			res.Errors = append(res.Errors, fmt.Sprintf("%s: %v", op.Purpose, err))
		}
	}
	s.log.Info("bank statement import finished",
		"total", len(ops),
		"imported", res.Imported,
		"skipped", res.Skipped,
		"unmatched", res.Unmatched,
		"errors", len(res.Errors),
	)
	return res, nil
}

// RawImportResult adds the chosen extractor name to the import result.
type RawImportResult struct {
	*ImportResult
	Extractor      string `json:"extractor"`       // "regex" or "llm:<model>"
	ExtractedCount int    `json:"extracted_count"` // operations the extractor produced
}

// ImportRawText extracts operations from raw statement text using the configured
// extractor (LLM if enabled, otherwise regex) and imports them idempotently.
func (s *ImportService) ImportRawText(ctx context.Context, rawText string) (*RawImportResult, error) {
	aiOps, err := s.extractor.Extract(ctx, rawText)
	if err != nil {
		s.log.Error("extraction failed", "extractor", s.extractor.Name(), "error", err)
		return nil, fmt.Errorf("extraction failed (%s): %w", s.extractor.Name(), err)
	}

	// Map ai.Operation -> BankOperation.
	ops := make([]BankOperation, 0, len(aiOps))
	for _, o := range aiOps {
		ops = append(ops, BankOperation{
			Date:          o.Date,
			Amount:        o.Amount,
			PayerINN:      o.PayerINN,
			PayerName:     o.PayerName,
			Purpose:       o.Purpose,
			InvoiceNumber: o.InvoiceNumber,
		})
	}

	res, _ := s.Import(ops)
	return &RawImportResult{
		ImportResult:   res,
		Extractor:      s.extractor.Name(),
		ExtractedCount: len(aiOps),
	}, nil
}

func (s *ImportService) importOne(op BankOperation, extID string, res *ImportResult) error {
	return retry(importMaxRetries, func() error {
		return s.db.Transaction(func(tx *gorm.DB) error {
			// Idempotency check: if a payment with this ExternalID exists, skip.
			var existing models.Payment
			err := tx.Where("external_id = ?", extID).First(&existing).Error
			if err == nil {
				res.Skipped++
				return nil
			}
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}

			// Match payer INN -> client. Lock the client row to avoid racing
			// with a concurrent import creating the same client.
			var client models.Client
			cErr := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
				Where("inn = ?", op.PayerINN).First(&client).Error
			if errors.Is(cErr, gorm.ErrRecordNotFound) {
				// Unknown payer: in a real system we'd create a draft client and
				// flag for manual review. Here we count it as unmatched.
				res.Unmatched++
				return nil
			}
			if cErr != nil {
				return cErr
			}

			// Naive project matching: first active project of the client.
			var project models.Project
			pErr := tx.Where("client_id = ?", client.ID).
				Order("created_at").First(&project).Error
			if pErr != nil {
				res.Unmatched++
				return nil
			}

			date, _ := time.Parse("2006-01-02", op.Date)

			payment := models.Payment{
				ProjectID:      project.ID,
				LegalEntityID:  client.ID,
				PaymentDate:    date,
				Amount:         op.Amount,
				PaymentPurpose: op.Purpose,
				ServiceStage:   guessStage(op.Purpose),
				InvoiceNumber:  op.InvoiceNumber,
				ExternalID:     extID,
			}
			if err := tx.Create(&payment).Error; err != nil {
				// A concurrent import may have inserted the same ExternalID
				// between our check and create -> treat unique violation as skip.
				if strings.Contains(strings.ToLower(err.Error()), "unique") {
					res.Skipped++
					return nil
				}
				return err
			}
			res.Imported++
			return nil
		})
	})
}

// guessStage maps keywords in the payment purpose to a service stage.
func guessStage(purpose string) string {
	p := strings.ToLower(purpose)
	switch {
	case strings.Contains(p, "seo"):
		return "SEO"
	case strings.Contains(p, "дизайн"):
		return "Дизайн"
	case strings.Contains(p, "реклам") || strings.Contains(p, "директ"):
		return "Реклама"
	case strings.Contains(p, "контент") || strings.Contains(p, "стат"):
		return "Контент"
	case strings.Contains(p, "сопровожд") || strings.Contains(p, "поддержк"):
		return "Сопровождение"
	default:
		return "Разработка"
	}
}

func retry(attempts int, fn func() error) error {
	var err error
	for i := 0; i < attempts; i++ {
		err = fn()
		if err == nil || !isLockConflict(err) {
			return err
		}
		time.Sleep(importRetryDelay * time.Duration(i+1))
	}
	return err
}
