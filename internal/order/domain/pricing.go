package domain

// bulkDiscountThreshold is the number of units at which the bulk discount
// applies.
const bulkDiscountThreshold = 10

// DiscountPolicy is a stateless domain service that decides the bulk discount
// for an order. It depends only on domain objects, so it belongs to the domain
// layer rather than to a use case.
type DiscountPolicy struct{}

// NewDiscountPolicy returns the default bulk-discount policy: 10% off the order
// total once the order contains bulkDiscountThreshold units or more.
func NewDiscountPolicy() DiscountPolicy { return DiscountPolicy{} }

// Discount returns the amount to subtract from the order total. It is the zero
// Money in the order currency when no discount applies.
func (DiscountPolicy) Discount(order *Order) (Money, error) {
	total, err := order.Total()
	if err != nil {
		return Money{}, err
	}
	if order.totalQuantity() < bulkDiscountThreshold {
		return ZeroMoney(total.Currency()), nil
	}
	return Money{amount: total.Amount() / 10, currency: total.Currency()}, nil
}
