package app_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/rzfhlv/go-std/internal/order/adapter/out/memory"
	"github.com/rzfhlv/go-std/internal/order/app"
	"github.com/rzfhlv/go-std/internal/order/domain"
)

type fixedClock struct {
	at time.Time
}

func (c fixedClock) Now() time.Time { return c.at }

type recordingPublisher struct {
	events []domain.Event
}

func (p *recordingPublisher) Publish(_ context.Context, events []domain.Event) error {
	p.events = append(p.events, events...)
	return nil
}

func newService(t *testing.T) (*app.OrderService, *recordingPublisher) {
	t.Helper()
	publisher := &recordingPublisher{}
	clock := fixedClock{at: time.Unix(1_700_000_000, 0).UTC()}
	return app.NewOrderService(memory.NewOrderRepository(), publisher, clock), publisher
}

func TestPlaceOrderPublishesEvents(t *testing.T) {
	t.Parallel()

	svc, publisher := newService(t)
	out, err := svc.PlaceOrder(context.Background(), app.PlaceOrderInput{
		OrderID:    "o1",
		CustomerID: "c1",
		Lines: []app.LineInput{
			{ProductID: "p1", Name: "Widget", UnitPrice: 500, Currency: "USD", Quantity: 2},
		},
	})
	if err != nil {
		t.Fatalf("PlaceOrder() error = %v", err)
	}
	if out.Status != string(domain.OrderStatusPending) {
		t.Errorf("Status = %q, want pending", out.Status)
	}
	if out.Total != 1000 || out.Currency != "USD" {
		t.Errorf("Total = %d %s, want 1000 USD", out.Total, out.Currency)
	}
	if len(publisher.events) != 2 || publisher.events[0].EventName() != "order.placed" {
		t.Fatalf("published events = %v, want [order.placed, order.line_added]", publisher.events)
	}
}

func TestConfirmOrderAppliesBulkDiscount(t *testing.T) {
	t.Parallel()

	svc, _ := newService(t)
	ctx := context.Background()
	if _, err := svc.PlaceOrder(ctx, app.PlaceOrderInput{
		OrderID:    "o1",
		CustomerID: "c1",
		Lines: []app.LineInput{
			{ProductID: "p1", Name: "Widget", UnitPrice: 100, Currency: "USD", Quantity: 10},
		},
	}); err != nil {
		t.Fatalf("PlaceOrder() error = %v", err)
	}

	out, err := svc.ConfirmOrder(ctx, "o1")
	if err != nil {
		t.Fatalf("ConfirmOrder() error = %v", err)
	}
	if out.Status != string(domain.OrderStatusConfirmed) {
		t.Errorf("Status = %q, want confirmed", out.Status)
	}
	if out.Subtotal != 1000 || out.Discount != 100 || out.Total != 900 {
		t.Errorf("pricing = sub %d disc %d total %d, want 1000/100/900", out.Subtotal, out.Discount, out.Total)
	}
}

func TestAddLineAfterConfirmFails(t *testing.T) {
	t.Parallel()

	svc, _ := newService(t)
	ctx := context.Background()
	if _, err := svc.PlaceOrder(ctx, app.PlaceOrderInput{
		OrderID:    "o1",
		CustomerID: "c1",
		Lines: []app.LineInput{
			{ProductID: "p1", Name: "Widget", UnitPrice: 500, Currency: "USD", Quantity: 1},
		},
	}); err != nil {
		t.Fatalf("PlaceOrder() error = %v", err)
	}
	if _, err := svc.ConfirmOrder(ctx, "o1"); err != nil {
		t.Fatalf("ConfirmOrder() error = %v", err)
	}

	_, err := svc.AddLine(ctx, "o1", app.LineInput{ProductID: "p2", Name: "Widget", UnitPrice: 100, Currency: "USD", Quantity: 1})
	if !errors.Is(err, domain.ErrInvalidStateTransition) {
		t.Fatalf("AddLine() error = %v, want ErrInvalidStateTransition", err)
	}
}

func TestGetOrderNotFound(t *testing.T) {
	t.Parallel()

	svc, _ := newService(t)
	if _, err := svc.GetOrder(context.Background(), "missing"); !errors.Is(err, domain.ErrOrderNotFound) {
		t.Fatalf("GetOrder() error = %v, want ErrOrderNotFound", err)
	}
}
