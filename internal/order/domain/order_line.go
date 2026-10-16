package domain

import (
	"fmt"
	"strings"
)

// OrderLine is an entity within the Order aggregate, identified by the product
// it references. All of its data is validated at construction.
type OrderLine struct {
	productID ProductID
	name      string
	unitPrice Money
	quantity  Quantity
}

// NewOrderLine constructs a validated line. It is used when adding lines to an
// order and when repository adapters reconstitute stored lines.
func NewOrderLine(productID ProductID, name string, unitPrice Money, quantity Quantity) (OrderLine, error) {
	trimmed := strings.TrimSpace(name)
	if productID == "" {
		return OrderLine{}, fmt.Errorf("%w: product id is required", ErrInvalidOrder)
	}
	if trimmed == "" {
		return OrderLine{}, fmt.Errorf("%w: product name is required", ErrInvalidOrder)
	}
	return OrderLine{productID: productID, name: trimmed, unitPrice: unitPrice, quantity: quantity}, nil
}

// ProductID returns the referenced product identifier.
func (l OrderLine) ProductID() ProductID { return l.productID }

// Name returns the product name captured on the line.
func (l OrderLine) Name() string { return l.name }

// UnitPrice returns the price of a single unit.
func (l OrderLine) UnitPrice() Money { return l.unitPrice }

// Quantity returns the number of units on the line.
func (l OrderLine) Quantity() Quantity { return l.quantity }

// Subtotal returns unitPrice multiplied by quantity.
func (l OrderLine) Subtotal() Money { return l.unitPrice.Multiply(l.quantity) }

func (l OrderLine) withQuantity(quantity Quantity) OrderLine {
	l.quantity = quantity
	return l
}
