// Package rest exposes the order module's use cases as a JSON REST API on top of
// the standard library net/http server.
package rest

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/rzfhlv/go-std/internal/order/app"
	"github.com/rzfhlv/go-std/internal/order/domain"
)

// Server adapts HTTP requests to the order use cases.
type Server struct {
	orders app.OrderUseCases
}

// NewServer builds a REST adapter around the given use cases.
func NewServer(orders app.OrderUseCases) *Server {
	return &Server{orders: orders}
}

// Register mounts the order routes onto mux so a command root can compose
// several modules onto one server.
func (s *Server) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /orders", s.listOrders)
	mux.HandleFunc("POST /orders", s.placeOrder)
	mux.HandleFunc("GET /orders/{id}", s.getOrder)
	mux.HandleFunc("POST /orders/{id}/lines", s.addLine)
	mux.HandleFunc("POST /orders/{id}/confirm", s.confirmOrder)
	mux.HandleFunc("POST /orders/{id}/cancel", s.cancelOrder)
}

// Handler returns a standalone handler with the order routes registered. It is
// convenient for tests.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	s.Register(mux)
	return mux
}

type lineRequest struct {
	ProductID string `json:"product_id"`
	Name      string `json:"name"`
	UnitPrice int64  `json:"unit_price"`
	Currency  string `json:"currency"`
	Quantity  int    `json:"quantity"`
}

func (l lineRequest) toInput() app.LineInput {
	return app.LineInput{
		ProductID: l.ProductID,
		Name:      l.Name,
		UnitPrice: l.UnitPrice,
		Currency:  l.Currency,
		Quantity:  l.Quantity,
	}
}

type placeOrderRequest struct {
	OrderID    string        `json:"order_id"`
	CustomerID string        `json:"customer_id"`
	Lines      []lineRequest `json:"lines"`
}

type cancelRequest struct {
	Reason string `json:"reason"`
}

type lineResponse struct {
	ProductID string `json:"product_id"`
	Name      string `json:"name"`
	UnitPrice int64  `json:"unit_price"`
	Quantity  int    `json:"quantity"`
	Subtotal  int64  `json:"subtotal"`
}

type orderResponse struct {
	ID         string         `json:"id"`
	CustomerID string         `json:"customer_id"`
	Status     string         `json:"status"`
	Lines      []lineResponse `json:"lines"`
	Subtotal   int64          `json:"subtotal"`
	Discount   int64          `json:"discount"`
	Total      int64          `json:"total"`
	Currency   string         `json:"currency"`
}

func toResponse(order app.OrderOutput) orderResponse {
	lines := make([]lineResponse, 0, len(order.Lines))
	for _, line := range order.Lines {
		lines = append(lines, lineResponse{
			ProductID: line.ProductID,
			Name:      line.Name,
			UnitPrice: line.UnitPrice,
			Quantity:  line.Quantity,
			Subtotal:  line.Subtotal,
		})
	}
	return orderResponse{
		ID:         order.ID,
		CustomerID: order.CustomerID,
		Status:     order.Status,
		Lines:      lines,
		Subtotal:   order.Subtotal,
		Discount:   order.Discount,
		Total:      order.Total,
		Currency:   order.Currency,
	}
}

func (s *Server) placeOrder(w http.ResponseWriter, r *http.Request) {
	var req placeOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	lines := make([]app.LineInput, 0, len(req.Lines))
	for _, line := range req.Lines {
		lines = append(lines, line.toInput())
	}

	out, err := s.orders.PlaceOrder(r.Context(), app.PlaceOrderInput{
		OrderID:    req.OrderID,
		CustomerID: req.CustomerID,
		Lines:      lines,
	})
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, toResponse(out))
}

func (s *Server) addLine(w http.ResponseWriter, r *http.Request) {
	var req lineRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	out, err := s.orders.AddLine(r.Context(), r.PathValue("id"), req.toInput())
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toResponse(out))
}

func (s *Server) confirmOrder(w http.ResponseWriter, r *http.Request) {
	out, err := s.orders.ConfirmOrder(r.Context(), r.PathValue("id"))
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toResponse(out))
}

func (s *Server) cancelOrder(w http.ResponseWriter, r *http.Request) {
	var req cancelRequest
	_ = json.NewDecoder(r.Body).Decode(&req)

	out, err := s.orders.CancelOrder(r.Context(), r.PathValue("id"), req.Reason)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toResponse(out))
}

func (s *Server) getOrder(w http.ResponseWriter, r *http.Request) {
	out, err := s.orders.GetOrder(r.Context(), r.PathValue("id"))
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toResponse(out))
}

func (s *Server) listOrders(w http.ResponseWriter, r *http.Request) {
	orders, err := s.orders.ListOrders(r.Context())
	if err != nil {
		writeDomainError(w, err)
		return
	}
	response := make([]orderResponse, 0, len(orders))
	for _, order := range orders {
		response = append(response, toResponse(order))
	}
	writeJSON(w, http.StatusOK, response)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	// The response is already committed; an encode error cannot be surfaced.
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func writeDomainError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrOrderNotFound):
		writeError(w, http.StatusNotFound, "order not found")
	case errors.Is(err, domain.ErrInvalidStateTransition):
		writeError(w, http.StatusConflict, err.Error())
	case errors.Is(err, domain.ErrInvalidOrder),
		errors.Is(err, domain.ErrEmptyOrder),
		errors.Is(err, domain.ErrInvalidQuantity),
		errors.Is(err, domain.ErrInvalidAmount),
		errors.Is(err, domain.ErrCurrencyMismatch):
		writeError(w, http.StatusBadRequest, err.Error())
	default:
		writeError(w, http.StatusInternalServerError, "internal error")
	}
}
