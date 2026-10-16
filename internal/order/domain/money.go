package domain

import (
	"fmt"
	"strings"
)

// Money is an immutable value object representing an amount in minor units (for
// example cents) together with its currency code. Two Money values are equal
// when both fields are equal.
type Money struct {
	amount   int64
	currency string
}

// NewMoney constructs a Money. The amount must not be negative and the currency
// must be non-empty; the currency code is upper-cased.
func NewMoney(amount int64, currency string) (Money, error) {
	if amount < 0 {
		return Money{}, fmt.Errorf("%w: %d", ErrInvalidAmount, amount)
	}
	currency = strings.ToUpper(strings.TrimSpace(currency))
	if currency == "" {
		return Money{}, fmt.Errorf("%w: currency is required", ErrInvalidAmount)
	}
	return Money{amount: amount, currency: currency}, nil
}

// ZeroMoney returns a zero Money in the given currency.
func ZeroMoney(currency string) Money {
	return Money{currency: strings.ToUpper(currency)}
}

// Amount returns the amount in minor units.
func (m Money) Amount() int64 { return m.amount }

// Currency returns the currency code.
func (m Money) Currency() string { return m.currency }

// IsZero reports whether the amount is zero.
func (m Money) IsZero() bool { return m.amount == 0 }

// Add returns the sum of two Money values of the same currency.
func (m Money) Add(other Money) (Money, error) {
	if m.currency != other.currency {
		return Money{}, fmt.Errorf("%w: %s and %s", ErrCurrencyMismatch, m.currency, other.currency)
	}
	return Money{amount: m.amount + other.amount, currency: m.currency}, nil
}

// Subtract returns m minus other, both of the same currency. The result must not
// be negative.
func (m Money) Subtract(other Money) (Money, error) {
	if m.currency != other.currency {
		return Money{}, fmt.Errorf("%w: %s and %s", ErrCurrencyMismatch, m.currency, other.currency)
	}
	if other.amount > m.amount {
		return Money{}, fmt.Errorf("%w: cannot subtract %d from %d", ErrInvalidAmount, other.amount, m.amount)
	}
	return Money{amount: m.amount - other.amount, currency: m.currency}, nil
}

// Multiply returns the Money scaled by a Quantity.
func (m Money) Multiply(quantity Quantity) Money {
	return Money{amount: m.amount * int64(quantity.value), currency: m.currency}
}

// String renders the Money as "amount currency".
func (m Money) String() string { return fmt.Sprintf("%d %s", m.amount, m.currency) }
