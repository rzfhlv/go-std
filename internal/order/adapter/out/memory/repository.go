// Package memory provides an in-memory implementation of the order module's
// outbound ports. It maps order snapshots back into domain aggregates at the
// boundary.
package memory

import (
	"context"
	"fmt"
	"sync"

	"github.com/rzfhlv/go-std/internal/order/domain"
	"github.com/rzfhlv/go-std/internal/order/port"
)

type lineSnapshot struct {
	productID string
	name      string
	unitPrice int64
	currency  string
	quantity  int
}

type orderSnapshot struct {
	id         domain.OrderID
	customerID domain.CustomerID
	status     domain.OrderStatus
	lines      []lineSnapshot
}

// OrderRepository stores orders in memory and is safe for concurrent use.
type OrderRepository struct {
	mu     sync.RWMutex
	orders map[domain.OrderID]orderSnapshot
}

// NewOrderRepository returns an empty in-memory repository.
func NewOrderRepository() *OrderRepository {
	return &OrderRepository{orders: make(map[domain.OrderID]orderSnapshot)}
}

var _ port.OrderRepository = (*OrderRepository)(nil)

// Save implements port.OrderRepository.
func (r *OrderRepository) Save(_ context.Context, order *domain.Order) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.orders[order.ID()] = toSnapshot(order)
	return nil
}

// FindByID implements port.OrderRepository.
func (r *OrderRepository) FindByID(_ context.Context, id domain.OrderID) (*domain.Order, error) {
	r.mu.RLock()
	snapshot, ok := r.orders[id]
	r.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("find order %s: %w", id, domain.ErrOrderNotFound)
	}
	return fromSnapshot(snapshot)
}

// List implements port.OrderRepository.
func (r *OrderRepository) List(_ context.Context) ([]*domain.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	orders := make([]*domain.Order, 0, len(r.orders))
	for _, snapshot := range r.orders {
		order, err := fromSnapshot(snapshot)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}
	return orders, nil
}

func toSnapshot(order *domain.Order) orderSnapshot {
	lines := order.Lines()
	snapshot := orderSnapshot{
		id:         order.ID(),
		customerID: order.CustomerID(),
		status:     order.Status(),
		lines:      make([]lineSnapshot, 0, len(lines)),
	}
	for _, line := range lines {
		snapshot.lines = append(snapshot.lines, lineSnapshot{
			productID: string(line.ProductID()),
			name:      line.Name(),
			unitPrice: line.UnitPrice().Amount(),
			currency:  line.UnitPrice().Currency(),
			quantity:  line.Quantity().Value(),
		})
	}
	return snapshot
}

func fromSnapshot(snapshot orderSnapshot) (*domain.Order, error) {
	lines := make([]domain.OrderLine, 0, len(snapshot.lines))
	for _, line := range snapshot.lines {
		unitPrice, err := domain.NewMoney(line.unitPrice, line.currency)
		if err != nil {
			return nil, fmt.Errorf("reconstitute order %s: %w", snapshot.id, err)
		}
		quantity, err := domain.NewQuantity(line.quantity)
		if err != nil {
			return nil, fmt.Errorf("reconstitute order %s: %w", snapshot.id, err)
		}
		reconstituted, err := domain.NewOrderLine(domain.ProductID(line.productID), line.name, unitPrice, quantity)
		if err != nil {
			return nil, fmt.Errorf("reconstitute order %s: %w", snapshot.id, err)
		}
		lines = append(lines, reconstituted)
	}
	order, err := domain.ReconstituteOrder(snapshot.id, snapshot.customerID, snapshot.status, lines)
	if err != nil {
		return nil, fmt.Errorf("reconstitute order %s: %w", snapshot.id, err)
	}
	return order, nil
}
