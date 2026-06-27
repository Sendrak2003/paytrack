package services

import (
	"io"
	"log/slog"
	"testing"
	"time"

	"payments-dashboard/internal/ai"
	"payments-dashboard/internal/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupImportDB(t *testing.T) (*ImportService, *gorm.DB) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&models.Client{}, &models.Project{}, &models.Payment{}, &models.Act{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	client := models.Client{Name: "Альфа", INN: "7701234567"}
	db.Create(&client)
	db.Create(&models.Project{Name: "Сайт", ClientID: client.ID, Status: "active"})
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	extractor := ai.NewExtractor(ai.Config{Provider: "", Log: log}) // regex fallback
	return NewImportService(db, log, extractor), db
}

func TestGuessStage(t *testing.T) {
	tests := []struct {
		purpose string
		want    string
	}{
		{"Оплата за SEO-продвижение", "SEO"},
		{"Оплата дизайна логотипа", "Дизайн"},
		{"Ведение Яндекс.Директ", "Реклама"},
		{"Написание статей (контент)", "Контент"},
		{"Техническое сопровождение", "Сопровождение"},
		{"Оплата по договору", "Разработка"},
	}
	for _, tt := range tests {
		t.Run(tt.purpose, func(t *testing.T) {
			if got := guessStage(tt.purpose); got != tt.want {
				t.Errorf("guessStage(%q) = %q, want %q", tt.purpose, got, tt.want)
			}
		})
	}
}

func TestExternalID_Deterministic(t *testing.T) {
	op := BankOperation{Date: "2024-10-05", Amount: 150000, PayerINN: "7701234567", Purpose: "аванс"}
	a := externalID(op)
	b := externalID(op)
	if a != b {
		t.Errorf("externalID not deterministic: %q != %q", a, b)
	}
	// Different amount -> different id.
	op2 := op
	op2.Amount = 150001
	if externalID(op2) == a {
		t.Error("externalID collision on different amount")
	}
}

func TestImport_Idempotent(t *testing.T) {
	svc, db := setupImportDB(t)

	ops := []BankOperation{
		{Date: "2024-10-05", Amount: 150000, PayerINN: "7701234567", Purpose: "Аванс за сайт", InvoiceNumber: "101"},
		{Date: "2024-11-01", Amount: 40000, PayerINN: "7701234567", Purpose: "SEO октябрь", InvoiceNumber: "119"},
	}

	// First import.
	res1, err := svc.Import(ops)
	if err != nil {
		t.Fatalf("import1: %v", err)
	}
	if res1.Imported != 2 {
		t.Errorf("first import: imported = %d, want 2", res1.Imported)
	}

	// Second import of the same batch — must skip everything.
	res2, err := svc.Import(ops)
	if err != nil {
		t.Fatalf("import2: %v", err)
	}
	if res2.Imported != 0 || res2.Skipped != 2 {
		t.Errorf("second import: imported=%d skipped=%d, want imported=0 skipped=2", res2.Imported, res2.Skipped)
	}

	var count int64
	db.Model(&models.Payment{}).Count(&count)
	if count != 2 {
		t.Errorf("expected 2 payments after double import, got %d", count)
	}
}

func TestImport_UnmatchedPayer(t *testing.T) {
	svc, _ := setupImportDB(t)

	ops := []BankOperation{
		{Date: "2024-10-05", Amount: 99999, PayerINN: "0000000000", Purpose: "Неизвестный плательщик"},
	}
	res, err := svc.Import(ops)
	if err != nil {
		t.Fatalf("import: %v", err)
	}
	if res.Unmatched != 1 || res.Imported != 0 {
		t.Errorf("unmatched=%d imported=%d, want unmatched=1 imported=0", res.Unmatched, res.Imported)
	}
}

func TestImport_StoresCorrectFields(t *testing.T) {
	svc, db := setupImportDB(t)
	_, err := svc.Import([]BankOperation{
		{Date: "2024-10-05", Amount: 150000, PayerINN: "7701234567", Purpose: "Аванс за SEO", InvoiceNumber: "101"},
	})
	if err != nil {
		t.Fatalf("import: %v", err)
	}
	var p models.Payment
	db.First(&p)
	if p.Amount != 150000 {
		t.Errorf("amount = %v, want 150000", p.Amount)
	}
	if p.ServiceStage != "SEO" {
		t.Errorf("service_stage = %q, want SEO", p.ServiceStage)
	}
	wantDate, _ := time.Parse("2006-01-02", "2024-10-05")
	if !p.PaymentDate.Equal(wantDate) {
		t.Errorf("payment_date = %v, want %v", p.PaymentDate, wantDate)
	}
	if p.ExternalID == "" {
		t.Error("external_id not set")
	}
}
