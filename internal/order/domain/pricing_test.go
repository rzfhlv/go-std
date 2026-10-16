package domain_test

import (
	"testing"

	"github.com/rzfhlv/go-std/internal/order/domain"
)

func TestDiscountPolicy(t *testing.T) {
	t.Parallel()

	policy := domain.NewDiscountPolicy()

	tests := []struct {
		name         string
		quantity     int
		wantDiscount int64
	}{
		{name: "below threshold has no discount", quantity: 9, wantDiscount: 0},
		{name: "at threshold gets ten percent", quantity: 10, wantDiscount: 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			order := mustOrder(t)
			addLine(t, order, "p1", 100, tt.quantity)

			discount, err := policy.Discount(order)
			if err != nil {
				t.Fatalf("Discount() error = %v", err)
			}
			if discount.Amount() != tt.wantDiscount {
				t.Errorf("Discount() = %d, want %d", discount.Amount(), tt.wantDiscount)
			}
		})
	}
}
