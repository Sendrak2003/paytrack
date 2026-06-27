package repository

import (
	"sync"
	"testing"
	"time"

	"payments-dashboard/internal/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupTestDB(t *testing.T) *Repository {
	t.Helper()
	// Shared in-memory DB for the duration of the test.
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&models.Client{}, &models.Project{}, &models.Payment{}, &models.Act{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	// Seed one client, project, payment.
	client := models.Client{Name: "Test Co", INN: "7700000000"}
	db.Create(&client)
	project := models.Project{Name: "Test Project", ClientID: client.ID, Status: "active"}
	db.Create(&project)
	payment := models.Payment{
		ProjectID:     project.ID,
		LegalEntityID: client.ID,
		PaymentDate:   time.Now().Add(-5 * 24 * time.Hour),
		Amount:        100000,
	}
	db.Create(&payment)
	return New(db)
}

func TestRepository_UpsertAct(t *testing.T) {
	tests := []struct {
		name       string
		isSent     bool
		isSigned   bool
		comment    string
		wantStatus string
		wantSentAt bool
	}{
		{name: "mark sent", isSent: true, isSigned: false, comment: "отправил", wantStatus: "waiting_signature", wantSentAt: true},
		{name: "mark sent and signed", isSent: true, isSigned: true, comment: "закрыт", wantStatus: "closed", wantSentAt: true},
		{name: "nothing set", isSent: false, isSigned: false, comment: "", wantStatus: "not_sent", wantSentAt: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := setupTestDB(t)
			act, err := repo.UpsertAct(1, tt.isSent, tt.isSigned, tt.comment)
			if err != nil {
				t.Fatalf("UpsertAct: %v", err)
			}
			if act.Status != tt.wantStatus {
				t.Errorf("status = %q, want %q", act.Status, tt.wantStatus)
			}
			if (act.SentAt != nil) != tt.wantSentAt {
				t.Errorf("SentAt present = %v, want %v", act.SentAt != nil, tt.wantSentAt)
			}
			if act.ManagerComment != tt.comment {
				t.Errorf("comment = %q, want %q", act.ManagerComment, tt.comment)
			}
		})
	}
}

// TestRepository_UpsertAct_Idempotent verifies that calling twice does not
// create a second act row and timestamps are stable across no-op updates.
func TestRepository_UpsertAct_Idempotent(t *testing.T) {
	repo := setupTestDB(t)

	first, err := repo.UpsertAct(1, true, false, "v1")
	if err != nil {
		t.Fatalf("first upsert: %v", err)
	}
	firstSentAt := *first.SentAt

	// Second call keeps is_sent=true; SentAt must not change.
	second, err := repo.UpsertAct(1, true, true, "v2")
	if err != nil {
		t.Fatalf("second upsert: %v", err)
	}
	if !second.SentAt.Equal(firstSentAt) {
		t.Errorf("SentAt changed on no-op transition: %v -> %v", firstSentAt, *second.SentAt)
	}
	if first.ID != second.ID {
		t.Errorf("act id changed: %d -> %d (duplicate row created)", first.ID, second.ID)
	}

	var count int64
	repo.db.Model(&models.Act{}).Where("payment_id = ?", 1).Count(&count)
	if count != 1 {
		t.Errorf("expected exactly 1 act row, got %d", count)
	}
}

// TestRepository_UpsertAct_Concurrent simulates two managers updating the same
// act simultaneously. With row locking + retry, both writes must succeed and
// exactly one row must exist.
func TestRepository_UpsertAct_Concurrent(t *testing.T) {
	repo := setupTestDB(t)

	var wg sync.WaitGroup
	errs := make([]error, 2)

	wg.Add(2)
	go func() {
		defer wg.Done()
		_, errs[0] = repo.UpsertAct(1, true, false, "manager A")
	}()
	go func() {
		defer wg.Done()
		_, errs[1] = repo.UpsertAct(1, true, true, "manager B")
	}()
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Errorf("goroutine %d failed: %v", i, err)
		}
	}

	var count int64
	repo.db.Model(&models.Act{}).Where("payment_id = ?", 1).Count(&count)
	if count != 1 {
		t.Errorf("expected exactly 1 act row after concurrent upserts, got %d", count)
	}
}

func TestIsLockConflict(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{name: "nil", err: nil, want: false},
		{name: "sqlite locked", err: errString("database is locked"), want: true},
		{name: "postgres serialization", err: errString("could not serialize access (SQLSTATE 40001)"), want: true},
		{name: "deadlock", err: errString("deadlock detected"), want: true},
		{name: "unrelated", err: errString("syntax error"), want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isLockConflict(tt.err); got != tt.want {
				t.Errorf("isLockConflict(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

type errString string

func (e errString) Error() string { return string(e) }
