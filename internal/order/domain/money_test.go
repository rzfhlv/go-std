package domain_test

import (
	"errors"
	"testing"

	"github.com/rzfhlv/go-std/internal/order/domain"
)

func TestNewMoney(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		amount   int64
		currency string
		wantErr  error
	}{
		{name: "valid", amount: 1000, currency: "USD"},
		{name: "upper-cases currency", amount: 100, currency: " usd "},
		{name: "negative amount", amount: -1, currency: "USD", wantErr: domain.ErrInvalidAmount},
		{name: "empty currency", amount: 1, currency: "   ", wantErr: domain.ErrInvalidAmount},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := domain.NewMoney(tt.amount, tt.currency)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("NewMoney() error = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("NewMoney() unexpected error: %v", err)
			}
			if got.Amount() != tt.amount {
				t.Errorf("Amount() = %d, want %d", got.Amount(), tt.amount)
			}
			if got.Currency() != "USD" {
				t.Errorf("Currency() = %q, want %q", got.Currency(), "USD")
			}
		})
	}
}

func TestMoneyArithmetic(t *testing.T) {
	t.Parallel()

	usd1000 := mustMoney(t, 1000, "USD")
	usd250 := mustMoney(t, 250, "USD")
	eur1 := mustMoney(t, 1, "EUR")

	sum, err := usd1000.Add(usd250)
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}
	if sum.Amount() != 1250 {
		t.Errorf("Add() = %d, want 1250", sum.Amount())
	}

	diff, err := usd1000.Subtract(usd250)
	if err != nil {
		t.Fatalf("Subtract() error = %v", err)
	}
	if diff.Amount() != 750 {
		t.Errorf("Subtract() = %d, want 750", diff.Amount())
	}

	if _, err := usd1000.Add(eur1); !errors.Is(err, domain.ErrCurrencyMismatch) {
		t.Errorf("Add(mismatch) error = %v, want ErrCurrencyMismatch", err)
	}
	if _, err := usd250.Subtract(usd1000); !errors.Is(err, domain.ErrInvalidAmount) {
		t.Errorf("Subtract(too large) error = %v, want ErrInvalidAmount", err)
	}
}

func mustMoney(t *testing.T, amount int64, currency string) domain.Money {
	t.Helper()
	money, err := domain.NewMoney(amount, currency)
	if err != nil {
		t.Fatalf("NewMoney(%d, %q): %v", amount, currency, err)
	}
	return money
}
