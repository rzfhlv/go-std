package domain

import "fmt"

// Quantity is an immutable value object representing a positive count of units.
type Quantity struct {
	value int
}

// NewQuantity constructs a Quantity, rejecting zero and negative values.
func NewQuantity(value int) (Quantity, error) {
	if value <= 0 {
		return Quantity{}, fmt.Errorf("%w: %d", ErrInvalidQuantity, value)
	}
	return Quantity{value: value}, nil
}

// Value returns the underlying count.
func (q Quantity) Value() int { return q.value }

// Add returns the sum of two quantities.
func (q Quantity) Add(other Quantity) Quantity {
	return Quantity{value: q.value + other.value}
}
