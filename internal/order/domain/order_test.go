package domain_test

import (
	"errors"
	"testing"
	"time"

	"github.com/rzfhlv/go-std/internal/order/domain"
)

var now = time.Unix(1_700_000_000, 0).UTC()

func mustOrder(t *testing.T) *domain.Order {
	t.Helper()
	order, err := domain.NewOrder("o1", "c1", now)
	if err != nil {
		t.Fatalf("NewOrder(): %v", err)
	}
	return order
}

func addLine(t *testing.T, order *domain.Order, product string, amount int64, quantity int) {
	t.Helper()
	if err := order.AddLine(domain.ProductID(product), "Widget", mustMoney(t, amount, "USD"), mustQuantity(t, quantity), now); err != nil {
		t.Fatalf("AddLine(%q): %v", product, err)
	}
}

func TestNewOrderValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		id         domain.OrderID
		customerID domain.CustomerID
		wantErr    error
	}{
		{name: "valid", id: "o1", customerID: "c1"},
		{name: "missing id", id: "", customerID: "c1", wantErr: domain.ErrInvalidOrder},
		{name: "missing customer", id: "o1", customerID: "", wantErr: domain.ErrInvalidOrder},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			order, err := domain.NewOrder(tt.id, tt.customerID, now)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("NewOrder() error = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("NewOrder() unexpected error: %v", err)
			}
			if order.Status() != domain.OrderStatusPending {
				t.Errorf("Status() = %q, want pending", order.Status())
			}
		})
	}
}

func TestOrderAddLineMergesSameProduct(t *testing.T) {
	t.Parallel()

	order := mustOrder(t)
	addLine(t, order, "p1", 500, 2)
	addLine(t, order, "p1", 500, 3)

	lines := order.Lines()
	if len(lines) != 1 {
		t.Fatalf("Lines() len = %d, want 1 (same product merges)", len(lines))
	}
	if got := lines[0].Quantity().Value(); got != 5 {
		t.Errorf("merged quantity = %d, want 5", got)
	}
}

func TestOrderTotal(t *testing.T) {
	t.Parallel()

	order := mustOrder(t)
	addLine(t, order, "p1", 500, 2)
	addLine(t, order, "p2", 1000, 1)

	total, err := order.Total()
	if err != nil {
		t.Fatalf("Total() error = %v", err)
	}
	if total.Amount() != 2000 || total.Currency() != "USD" {
		t.Errorf("Total() = %s, want 2000 USD", total)
	}
}

func TestOrderCurrencyConsistency(t *testing.T) {
	t.Parallel()

	order := mustOrder(t)
	addLine(t, order, "p1", 500, 1)

	eur, err := domain.NewMoney(500, "EUR")
	if err != nil {
		t.Fatalf("NewMoney: %v", err)
	}
	err = order.AddLine("p2", "Widget", eur, mustQuantity(t, 1), now)
	if !errors.Is(err, domain.ErrCurrencyMismatch) {
		t.Fatalf("AddLine(EUR into USD order) error = %v, want ErrCurrencyMismatch", err)
	}
}

func TestOrderConfirm(t *testing.T) {
	t.Parallel()

	t.Run("empty order cannot be confirmed", func(t *testing.T) {
		t.Parallel()

		order := mustOrder(t)
		if err := order.Confirm(now); !errors.Is(err, domain.ErrEmptyOrder) {
			t.Fatalf("Confirm() error = %v, want ErrEmptyOrder", err)
		}
	})

	t.Run("confirmed order records event and freezes lines", func(t *testing.T) {
		t.Parallel()

		order := mustOrder(t)
		addLine(t, order, "p1", 500, 2)
		order.PullEvents() // drop creation/add events

		if err := order.Confirm(now); err != nil {
			t.Fatalf("Confirm() error = %v", err)
		}
		if order.Status() != domain.OrderStatusConfirmed {
			t.Fatalf("Status() = %q, want confirmed", order.Status())
		}

		events := order.PullEvents()
		if len(events) != 1 || events[0].EventName() != "order.confirmed" {
			t.Fatalf("events = %v, want one order.confirmed", events)
		}
		confirmed, ok := events[0].(domain.OrderConfirmed)
		if !ok {
			t.Fatalf("event type = %T, want domain.OrderConfirmed", events[0])
		}
		if confirmed.Total != 1000 {
			t.Errorf("OrderConfirmed.Total = %d, want 1000", confirmed.Total)
		}

		err := order.AddLine("p9", "Widget", mustMoney(t, 100, "USD"), mustQuantity(t, 1), now)
		if !errors.Is(err, domain.ErrInvalidStateTransition) {
			t.Errorf("AddLine after confirm error = %v, want ErrInvalidStateTransition", err)
		}
	})

	t.Run("cannot confirm twice", func(t *testing.T) {
		t.Parallel()

		order := mustOrder(t)
		addLine(t, order, "p1", 500, 1)
		if err := order.Confirm(now); err != nil {
			t.Fatalf("first Confirm() error = %v", err)
		}
		if err := order.Confirm(now); !errors.Is(err, domain.ErrInvalidStateTransition) {
			t.Fatalf("second Confirm() error = %v, want ErrInvalidStateTransition", err)
		}
	})
}

func TestOrderCancel(t *testing.T) {
	t.Parallel()

	t.Run("pending order can be cancelled", func(t *testing.T) {
		t.Parallel()

		order := mustOrder(t)
		order.PullEvents()
		if err := order.Cancel("customer changed mind", now); err != nil {
			t.Fatalf("Cancel() error = %v", err)
		}
		if order.Status() != domain.OrderStatusCancelled {
			t.Fatalf("Status() = %q, want cancelled", order.Status())
		}
		events := order.PullEvents()
		if len(events) != 1 || events[0].EventName() != "order.cancelled" {
			t.Fatalf("events = %v, want one order.cancelled", events)
		}
	})

	t.Run("confirmed order cannot be cancelled", func(t *testing.T) {
		t.Parallel()

		order := mustOrder(t)
		addLine(t, order, "p1", 500, 1)
		if err := order.Confirm(now); err != nil {
			t.Fatalf("Confirm() error = %v", err)
		}
		if err := order.Cancel("too late", now); !errors.Is(err, domain.ErrInvalidStateTransition) {
			t.Fatalf("Cancel() error = %v, want ErrInvalidStateTransition", err)
		}
	})
}

func TestOrderPullEventsClears(t *testing.T) {
	t.Parallel()

	order := mustOrder(t)
	if got := len(order.PullEvents()); got != 1 {
		t.Fatalf("first PullEvents() len = %d, want 1 (order.placed)", got)
	}
	if got := len(order.PullEvents()); got != 0 {
		t.Fatalf("second PullEvents() len = %d, want 0", got)
	}
}
