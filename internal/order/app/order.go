package app

import (
	"context"
	"fmt"
	"time"

	"github.com/rzfhlv/go-std/internal/order/domain"
	"github.com/rzfhlv/go-std/internal/order/port"
)

// LineInput is the input DTO describing one order line.
type LineInput struct {
	ProductID string
	Name      string
	UnitPrice int64
	Currency  string
	Quantity  int
}

// PlaceOrderInput is the input DTO for placing an order.
type PlaceOrderInput struct {
	OrderID    string
	CustomerID string
	Lines      []LineInput
}

// LineOutput is the output DTO for an order line.
type LineOutput struct {
	ProductID string
	Name      string
	UnitPrice int64
	Quantity  int
	Subtotal  int64
}

// OrderOutput is the output DTO for an order, including pricing.
type OrderOutput struct {
	ID         string
	CustomerID string
	Status     string
	Lines      []LineOutput
	Subtotal   int64
	Discount   int64
	Total      int64
	Currency   string
}

// OrderUseCases is the inbound port implemented by OrderService. Every
// transport adapter drives these use cases.
type OrderUseCases interface {
	PlaceOrder(ctx context.Context, in PlaceOrderInput) (OrderOutput, error)
	AddLine(ctx context.Context, orderID string, in LineInput) (OrderOutput, error)
	ConfirmOrder(ctx context.Context, orderID string) (OrderOutput, error)
	CancelOrder(ctx context.Context, orderID, reason string) (OrderOutput, error)
	GetOrder(ctx context.Context, orderID string) (OrderOutput, error)
	ListOrders(ctx context.Context) ([]OrderOutput, error)
}

// OrderService orchestrates the ordering use cases through outbound ports. It
// holds no business rules of its own.
type OrderService struct {
	repo    port.OrderRepository
	events  port.EventPublisher
	clock   port.Clock
	pricing domain.DiscountPolicy
}

// NewOrderService wires an OrderService to its outbound ports.
func NewOrderService(repo port.OrderRepository, events port.EventPublisher, clock port.Clock) *OrderService {
	return &OrderService{repo: repo, events: events, clock: clock, pricing: domain.NewDiscountPolicy()}
}

var _ OrderUseCases = (*OrderService)(nil)

// PlaceOrder creates a pending order with the supplied lines and publishes the
// resulting domain events.
func (s *OrderService) PlaceOrder(ctx context.Context, in PlaceOrderInput) (OrderOutput, error) {
	order, err := domain.NewOrder(domain.OrderID(in.OrderID), domain.CustomerID(in.CustomerID), s.clock.Now())
	if err != nil {
		return OrderOutput{}, fmt.Errorf("place order: %w", err)
	}
	for _, line := range in.Lines {
		if err := addLine(order, line, s.clock.Now()); err != nil {
			return OrderOutput{}, fmt.Errorf("place order: %w", err)
		}
	}
	if err := s.persist(ctx, order); err != nil {
		return OrderOutput{}, err
	}
	return s.toOutput(order)
}

// AddLine adds a line to an existing pending order.
func (s *OrderService) AddLine(ctx context.Context, orderID string, in LineInput) (OrderOutput, error) {
	order, err := s.load(ctx, orderID)
	if err != nil {
		return OrderOutput{}, err
	}
	if err := addLine(order, in, s.clock.Now()); err != nil {
		return OrderOutput{}, fmt.Errorf("add line: %w", err)
	}
	if err := s.persist(ctx, order); err != nil {
		return OrderOutput{}, err
	}
	return s.toOutput(order)
}

// ConfirmOrder confirms a pending order with at least one line.
func (s *OrderService) ConfirmOrder(ctx context.Context, orderID string) (OrderOutput, error) {
	order, err := s.load(ctx, orderID)
	if err != nil {
		return OrderOutput{}, err
	}
	if err := order.Confirm(s.clock.Now()); err != nil {
		return OrderOutput{}, fmt.Errorf("confirm order: %w", err)
	}
	if err := s.persist(ctx, order); err != nil {
		return OrderOutput{}, err
	}
	return s.toOutput(order)
}

// CancelOrder cancels a pending order.
func (s *OrderService) CancelOrder(ctx context.Context, orderID, reason string) (OrderOutput, error) {
	order, err := s.load(ctx, orderID)
	if err != nil {
		return OrderOutput{}, err
	}
	if err := order.Cancel(reason, s.clock.Now()); err != nil {
		return OrderOutput{}, fmt.Errorf("cancel order: %w", err)
	}
	if err := s.persist(ctx, order); err != nil {
		return OrderOutput{}, err
	}
	return s.toOutput(order)
}

// GetOrder returns a single order by ID.
func (s *OrderService) GetOrder(ctx context.Context, orderID string) (OrderOutput, error) {
	order, err := s.load(ctx, orderID)
	if err != nil {
		return OrderOutput{}, err
	}
	return s.toOutput(order)
}

// ListOrders returns every stored order.
func (s *OrderService) ListOrders(ctx context.Context) ([]OrderOutput, error) {
	orders, err := s.repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list orders: %w", err)
	}
	out := make([]OrderOutput, 0, len(orders))
	for _, order := range orders {
		item, err := s.toOutput(order)
		if err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, nil
}

func addLine(order *domain.Order, in LineInput, at time.Time) error {
	unitPrice, err := domain.NewMoney(in.UnitPrice, in.Currency)
	if err != nil {
		return err
	}
	quantity, err := domain.NewQuantity(in.Quantity)
	if err != nil {
		return err
	}
	return order.AddLine(domain.ProductID(in.ProductID), in.Name, unitPrice, quantity, at)
}

func (s *OrderService) load(ctx context.Context, orderID string) (*domain.Order, error) {
	order, err := s.repo.FindByID(ctx, domain.OrderID(orderID))
	if err != nil {
		return nil, fmt.Errorf("load order %s: %w", orderID, err)
	}
	return order, nil
}

func (s *OrderService) persist(ctx context.Context, order *domain.Order) error {
	if err := s.repo.Save(ctx, order); err != nil {
		return fmt.Errorf("save order %s: %w", order.ID(), err)
	}
	events := order.PullEvents()
	if len(events) == 0 {
		return nil
	}
	if err := s.events.Publish(ctx, events); err != nil {
		return fmt.Errorf("publish order events: %w", err)
	}
	return nil
}

func (s *OrderService) toOutput(order *domain.Order) (OrderOutput, error) {
	subtotal, err := order.Total()
	if err != nil {
		return OrderOutput{}, err
	}
	discount, err := s.pricing.Discount(order)
	if err != nil {
		return OrderOutput{}, err
	}
	total, err := subtotal.Subtract(discount)
	if err != nil {
		return OrderOutput{}, err
	}

	lines := order.Lines()
	out := OrderOutput{
		ID:         string(order.ID()),
		CustomerID: string(order.CustomerID()),
		Status:     string(order.Status()),
		Subtotal:   subtotal.Amount(),
		Discount:   discount.Amount(),
		Total:      total.Amount(),
		Currency:   subtotal.Currency(),
		Lines:      make([]LineOutput, 0, len(lines)),
	}
	for _, line := range lines {
		out.Lines = append(out.Lines, LineOutput{
			ProductID: string(line.ProductID()),
			Name:      line.Name(),
			UnitPrice: line.UnitPrice().Amount(),
			Quantity:  line.Quantity().Value(),
			Subtotal:  line.Subtotal().Amount(),
		})
	}
	return out, nil
}
