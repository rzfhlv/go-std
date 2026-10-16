package domain_test

import (
	"errors"
	"testing"

	"github.com/rzfhlv/go-std/internal/order/domain"
)

func TestNewQuantity(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		value   int
		wantErr error
	}{
		{name: "valid", value: 1},
		{name: "zero", value: 0, wantErr: domain.ErrInvalidQuantity},
		{name: "negative", value: -3, wantErr: domain.ErrInvalidQuantity},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			quantity, err := domain.NewQuantity(tt.value)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("NewQuantity() error = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("NewQuantity() unexpected error: %v", err)
			}
			if quantity.Value() != tt.value {
				t.Errorf("Value() = %d, want %d", quantity.Value(), tt.value)
			}
		})
	}
}

func TestQuantityAdd(t *testing.T) {
	t.Parallel()

	one := mustQuantity(t, 1)
	two := mustQuantity(t, 2)
	if got := one.Add(two).Value(); got != 3 {
		t.Errorf("Add() = %d, want 3", got)
	}
}

func mustQuantity(t *testing.T, value int) domain.Quantity {
	t.Helper()
	quantity, err := domain.NewQuantity(value)
	if err != nil {
		t.Fatalf("NewQuantity(%d): %v", value, err)
	}
	return quantity
}
