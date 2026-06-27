package domain

import "context"

// BankOperation — одна операция из банковской выписки.
type BankOperation struct {
	Date          string
	Amount        float64
	PayerINN      string
	PayerName     string
	Purpose       string
	InvoiceNumber string
}

// ImportResult — результат импорта выписки.
type ImportResult struct {
	Imported  int
	Skipped   int
	Unmatched int
	Errors    []string
}

// RawImportResult — результат импорта из сырого текста.
type RawImportResult struct {
	*ImportResult
	Extractor      string
	ExtractedCount int
}

// Extractor — порт для извлечения операций из текста выписки.
type Extractor interface {
	Extract(ctx context.Context, rawText string) ([]BankOperation, error)
	Name() string
}

// ImportService — порт для сервиса импорта.
type ImportService interface {
	Import(ops []BankOperation) (*ImportResult, error)
	ImportRawText(ctx context.Context, rawText string) (*RawImportResult, error)
}
