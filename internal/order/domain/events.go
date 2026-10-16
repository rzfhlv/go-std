package domain

import "time"

// Event is a domain event emitted by the Order aggregate. The application layer
// and adapters treat it as opaque data.
type Event interface {
	// EventName returns the stable name of the event.
	EventName() string
	// OccurredAt returns when the event happened.
	OccurredAt() time.Time
}

// baseEvent carries the occurrence time shared by every domain event.
type baseEvent struct {
	at time.Time
}

func newBaseEvent(at time.Time) baseEvent { return baseEvent{at: at} }

// OccurredAt returns when the event happened.
func (e baseEvent) OccurredAt() time.Time { return e.at }

// OrderPlaced is emitted when an order is created.
type OrderPlaced struct {
	baseEvent
	OrderID    OrderID
	CustomerID CustomerID
}

// EventName returns the stable event name.
func (OrderPlaced) EventName() string { return "order.placed" }

// LineAdded is emitted when a line is added to an order.
type LineAdded struct {
	baseEvent
	OrderID   OrderID
	ProductID ProductID
	Quantity  int
}

// EventName returns the stable event name.
func (LineAdded) EventName() string { return "order.line_added" }

// OrderConfirmed is emitted when an order is confirmed, carrying its total.
type OrderConfirmed struct {
	baseEvent
	OrderID  OrderID
	Total    int64
	Currency string
}

// EventName returns the stable event name.
func (OrderConfirmed) EventName() string { return "order.confirmed" }

// OrderCancelled is emitted when an order is cancelled.
type OrderCancelled struct {
	baseEvent
	OrderID OrderID
	Reason  string
}

// EventName returns the stable event name.
func (OrderCancelled) EventName() string { return "order.cancelled" }
