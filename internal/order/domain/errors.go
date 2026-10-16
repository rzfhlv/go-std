package domain

import "errors"

// Domain sentinel errors. Adapters map these to transport errors with errors.Is.
var (
	// ErrInvalidOrder is returned when an order breaks a structural invariant.
	ErrInvalidOrder = errors.New("invalid order")
	// ErrEmptyOrder is returned when an operation requires at least one line.
	ErrEmptyOrder = errors.New("order has no lines")
	// ErrInvalidStateTransition is returned on an illegal status transition.
	ErrInvalidStateTransition = errors.New("invalid order state transition")
	// ErrOrderNotFound is returned when an order cannot be located.
	ErrOrderNotFound = errors.New("order not found")
	// ErrInvalidQuantity is returned when a quantity is not positive.
	ErrInvalidQuantity = errors.New("quantity must be greater than zero")
	// ErrInvalidAmount is returned when a monetary amount is negative or invalid.
	ErrInvalidAmount = errors.New("invalid amount")
	// ErrCurrencyMismatch is returned when money of different currencies is combined.
	ErrCurrencyMismatch = errors.New("currency mismatch")
)
