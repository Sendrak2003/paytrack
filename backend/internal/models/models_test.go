package models

import (
	"testing"
	"time"
)

func TestAct_CalculateStatus(t *testing.T) {
	now := time.Now()
	day := 24 * time.Hour

	ago := func(d time.Duration) *time.Time {
		t := now.Add(-d)
		return &t
	}

	tests := []struct {
		name        string
		act         Act
		paymentDate time.Time
		want        string
	}{
		{
			name:        "new payment, no act activity -> not_sent",
			act:         Act{IsSent: false, IsSigned: false},
			paymentDate: now.Add(-5 * day),
			want:        "not_sent",
		},
		{
			name:        "old payment (>30d), not sent -> needs_attention",
			act:         Act{IsSent: false, IsSigned: false},
			paymentDate: now.Add(-40 * day),
			want:        "needs_attention",
		},
		{
			name:        "sent recently, not signed -> waiting_signature",
			act:         Act{IsSent: true, SentAt: ago(3 * day), IsSigned: false},
			paymentDate: now.Add(-5 * day),
			want:        "waiting_signature",
		},
		{
			name:        "sent long ago (>14d), not signed -> needs_attention",
			act:         Act{IsSent: true, SentAt: ago(20 * day), IsSigned: false},
			paymentDate: now.Add(-25 * day),
			want:        "needs_attention",
		},
		{
			name:        "sent and signed -> closed",
			act:         Act{IsSent: true, SentAt: ago(10 * day), IsSigned: true, SignedAt: ago(2 * day)},
			paymentDate: now.Add(-15 * day),
			want:        "closed",
		},
		{
			name:        "signed implies closed even if sent recently",
			act:         Act{IsSent: true, SentAt: ago(1 * day), IsSigned: true, SignedAt: ago(1 * day)},
			paymentDate: now.Add(-2 * day),
			want:        "closed",
		},
		{
			name:        "sent exactly at boundary, not signed -> waiting_signature",
			act:         Act{IsSent: true, SentAt: ago(13 * day), IsSigned: false},
			paymentDate: now.Add(-15 * day),
			want:        "waiting_signature",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.act.CalculateStatus(tt.paymentDate)
			if got != tt.want {
				t.Errorf("CalculateStatus() = %q, want %q", got, tt.want)
			}
		})
	}
}
