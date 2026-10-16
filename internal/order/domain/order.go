package domain

import (
	"fmt"
	"time"
)

// OrderStatus is the lifecycle state of an Order.
type OrderStatus string

// Order lifecycle states.
const (
	// OrderStatusPending is the initial state; lines may be added.
	OrderStatusPending OrderStatus = "pending"
	// OrderStatusConfirmed is a confirmed order; its lines are frozen.
	OrderStatusConfirmed OrderStatus = "confirmed"
	// OrderStatusCancelled is a terminal state reachable only from pending.
	OrderStatusCancelled OrderStatus = "cancelled"
)

// Order is the aggregate root of the ordering bounded context. It guards every
// invariant of the consistency boundary: valid line data, a single currency
// across lines, and legal status transitions. Its state is only reachable
// through behavior, and it records the domain events that describe what
// happened.
type Order struct {
	id         OrderID
	customerID CustomerID
	lines      []OrderLine
	status     OrderStatus
	events     []Event
}

// NewOrder creates a pending order and records an OrderPlaced event.
func NewOrder(id OrderID, customerID CustomerID, at time.Time) (*Order, error) {
	if id == "" {
		return nil, fmt.Errorf("%w: order id is required", ErrInvalidOrder)
	}
	if customerID == "" {
		return nil, fmt.Errorf("%w: customer id is required", ErrInvalidOrder)
	}
	order := &Order{id: id, customerID: customerID, status: OrderStatusPending}
	order.record(OrderPlaced{baseEvent: newBaseEvent(at), OrderID: id, CustomerID: customerID})
	return order, nil
}

// ReconstituteOrder rebuilds an order from persisted state without emitting
// domain events. Repository adapters use it to hydrate aggregates.
func ReconstituteOrder(id OrderID, customerID CustomerID, status OrderStatus, lines []OrderLine) (*Order, error) {
	if id == "" || customerID == "" {
		return nil, fmt.Errorf("%w: order id and customer id are required", ErrInvalidOrder)
	}
	switch status {
	case OrderStatusPending, OrderStatusConfirmed, OrderStatusCancelled:
	default:
		return nil, fmt.Errorf("%w: unknown status %q", ErrInvalidOrder, status)
	}
	if err := ensureSingleCurrency(lines); err != nil {
		return nil, err
	}
	return &Order{id: id, customerID: customerID, status: status, lines: append([]OrderLine(nil), lines...)}, nil
}

func ensureSingleCurrency(lines []OrderLine) error {
	if len(lines) < 2 {
		return nil
	}
	currency := lines[0].unitPrice.Currency()
	for _, line := range lines[1:] {
		if line.unitPrice.Currency() != currency {
			return fmt.Errorf("%w: order currency is %s", ErrCurrencyMismatch, currency)
		}
	}
	return nil
}

// ID returns the order identifier.
func (o *Order) ID() OrderID { return o.id }

// CustomerID returns the customer that placed the order.
func (o *Order) CustomerID() CustomerID { return o.customerID }

// Status returns the current lifecycle state.
func (o *Order) Status() OrderStatus { return o.status }

// Lines returns a defensive copy of the order lines.
func (o *Order) Lines() []OrderLine { return append([]OrderLine(nil), o.lines...) }

// AddLine adds a line to a pending order. Adding a product that is already on
// the order merges the quantities instead of creating a duplicate line.
func (o *Order) AddLine(productID ProductID, name string, unitPrice Money, quantity Quantity, at time.Time) error {
	if o.status != OrderStatusPending {
		return fmt.Errorf("%w: cannot add a line to a %s order", ErrInvalidStateTransition, o.status)
	}
	if len(o.lines) > 0 && o.lines[0].unitPrice.Currency() != unitPrice.Currency() {
		return fmt.Errorf("%w: order currency is %s", ErrCurrencyMismatch, o.lines[0].unitPrice.Currency())
	}

	for i, line := range o.lines {
		if line.ProductID() == productID {
			o.lines[i] = line.withQuantity(line.Quantity().Add(quantity))
			o.record(LineAdded{baseEvent: newBaseEvent(at), OrderID: o.id, ProductID: productID, Quantity: quantity.Value()})
			return nil
		}
	}

	newLine, err := NewOrderLine(productID, name, unitPrice, quantity)
	if err != nil {
		return err
	}
	o.lines = append(o.lines, newLine)
	o.record(LineAdded{baseEvent: newBaseEvent(at), OrderID: o.id, ProductID: productID, Quantity: quantity.Value()})
	return nil
}

// Total returns the sum of all line subtotals, or the zero Money when the order
// has no lines.
func (o *Order) Total() (Money, error) {
	if len(o.lines) == 0 {
		return Money{}, nil
	}
	total := ZeroMoney(o.lines[0].unitPrice.Currency())
	for _, line := range o.lines {
		sum, err := total.Add(line.Subtotal())
		if err != nil {
			return Money{}, err
		}
		total = sum
	}
	return total, nil
}

// Confirm moves a pending order with at least one line to the confirmed state
// and records an OrderConfirmed event carrying the order total.
func (o *Order) Confirm(at time.Time) error {
	if o.status != OrderStatusPending {
		return fmt.Errorf("%w: cannot confirm a %s order", ErrInvalidStateTransition, o.status)
	}
	if len(o.lines) == 0 {
		return fmt.Errorf("%w: cannot confirm", ErrEmptyOrder)
	}
	total, err := o.Total()
	if err != nil {
		return err
	}
	o.status = OrderStatusConfirmed
	o.record(OrderConfirmed{baseEvent: newBaseEvent(at), OrderID: o.id, Total: total.Amount(), Currency: total.Currency()})
	return nil
}

// Cancel moves a pending order to the cancelled state and records an
// OrderCancelled event. Confirmed orders cannot be cancelled.
func (o *Order) Cancel(reason string, at time.Time) error {
	if o.status != OrderStatusPending {
		return fmt.Errorf("%w: cannot cancel a %s order", ErrInvalidStateTransition, o.status)
	}
	o.status = OrderStatusCancelled
	o.record(OrderCancelled{baseEvent: newBaseEvent(at), OrderID: o.id, Reason: reason})
	return nil
}

// PullEvents returns and clears the aggregate's pending domain events. The
// application layer calls it after persisting to publish what happened.
func (o *Order) PullEvents() []Event {
	events := o.events
	o.events = nil
	return events
}

func (o *Order) record(event Event) {
	o.events = append(o.events, event)
}

func (o *Order) totalQuantity() int {
	total := 0
	for _, line := range o.lines {
		total += line.quantity.value
	}
	return total
}
