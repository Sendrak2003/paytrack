package ai

import (
	"context"
	"log/slog"
	"testing"
)

func TestNewExtractor_FallbackToRegex(t *testing.T) {
	tests := []struct {
		provider string
		want     string
	}{
		{"", "regex"},
		{"unknown", "regex"},
	}
	for _, tt := range tests {
		t.Run(tt.provider, func(t *testing.T) {
			e := NewExtractor(Config{Provider: tt.provider, Log: slog.Default()})
			if e.Name() != tt.want {
				t.Errorf("Name() = %q, want %q", e.Name(), tt.want)
			}
		})
	}
}

func TestRegexExtractor_Extract(t *testing.T) {
	raw := `05.10.2024  Поступление  150 000,00 RUB
Плательщик: ООО «Альфа Медиа», ИНН 7701234567
Назначение: Оплата по счёту №101 за разработку сайта, аванс 50%

01.11.2024  Поступление  40 000,00 RUB
Плательщик: ООО «Альфа Медиа», ИНН 7701234567
Назначение: Оплата по счету № 119 за SEO октябрь

Остаток на счёте: 999 999,00 RUB`

	e := &regexExtractor{}
	ops, err := e.Extract(context.Background(), raw)
	if err != nil {
		t.Fatalf("Extract: %v", err)
	}

	// Two payment blocks; the "Остаток" block has no amount+date pair shaped as a payment? It does have both.
	// It will be filtered only if amount/date missing. "Остаток" has amount but the date regex won't match -> skipped.
	if len(ops) != 2 {
		t.Fatalf("expected 2 operations, got %d: %+v", len(ops), ops)
	}

	first := ops[0]
	if first.Date != "2024-10-05" {
		t.Errorf("date = %q, want 2024-10-05", first.Date)
	}
	if first.Amount != 150000.00 {
		t.Errorf("amount = %v, want 150000", first.Amount)
	}
	if first.PayerINN != "7701234567" {
		t.Errorf("inn = %q, want 7701234567", first.PayerINN)
	}
	if first.InvoiceNumber != "101" {
		t.Errorf("invoice = %q, want 101", first.InvoiceNumber)
	}

	if ops[1].InvoiceNumber != "119" {
		t.Errorf("second invoice = %q, want 119", ops[1].InvoiceNumber)
	}
}

func TestStripCodeFence(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{`{"a":1}`, `{"a":1}`},
		{"```json\n{\"a\":1}\n```", `{"a":1}`},
		{"```\n{\"a\":1}\n```", `{"a":1}`},
		{"  {\"a\":1}  ", `{"a":1}`},
	}
	for _, tt := range tests {
		if got := stripCodeFence(tt.in); got != tt.want {
			t.Errorf("stripCodeFence(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
