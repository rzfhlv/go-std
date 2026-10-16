// Package port defines the ordering module's outbound interfaces (what the
// module needs from infrastructure). It contains interfaces only.
package port

import (
	"context"
	"time"

	"github.com/rzfhlv/go-std/internal/order/domain"
)

// OrderRepository is the outbound port for persisting Order aggregates.
type OrderRepository interface {
	// Save persists an order, overwriting any entry with the same ID.
	Save(ctx context.Context, order *domain.Order) error
	// FindByID returns the order with the given ID or domain.ErrOrderNotFound.
	FindByID(ctx context.Context, id domain.OrderID) (*domain.Order, error)
	// List returns every stored order.
	List(ctx context.Context) ([]*domain.Order, error)
}

// EventPublisher is the outbound port for publishing the domain events raised by
// an aggregate.
type EventPublisher interface {
	Publish(ctx context.Context, events []domain.Event) error
}

// Clock is the outbound port supplying the current time, so use cases stay
// deterministic in tests.
type Clock interface {
	Now() time.Time
}
