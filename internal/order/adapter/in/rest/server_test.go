package rest_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/rzfhlv/go-std/internal/order/adapter/in/rest"
	"github.com/rzfhlv/go-std/internal/order/adapter/out/memory"
	"github.com/rzfhlv/go-std/internal/order/app"
	"github.com/rzfhlv/go-std/internal/order/domain"
)

type fixedClock struct {
	at time.Time
}

func (c fixedClock) Now() time.Time { return c.at }

type discardPublisher struct{}

func (discardPublisher) Publish(context.Context, []domain.Event) error { return nil }

func newHandler() http.Handler {
	service := app.NewOrderService(memory.NewOrderRepository(), discardPublisher{}, fixedClock{at: time.Unix(0, 0).UTC()})
	return rest.NewServer(service).Handler()
}

func TestPlaceThenConfirmOrder(t *testing.T) {
	t.Parallel()

	h := newHandler()

	body := `{"order_id":"o1","customer_id":"c1","lines":[{"product_id":"p1","name":"Widget","unit_price":100,"currency":"USD","quantity":10}]}`
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/orders", strings.NewReader(body))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("POST /orders status = %d, want 201; body=%s", rec.Code, rec.Body.String())
	}

	req = httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/orders/o1/confirm", nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("POST /orders/o1/confirm status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}

	var got struct {
		Status   string `json:"status"`
		Subtotal int64  `json:"subtotal"`
		Discount int64  `json:"discount"`
		Total    int64  `json:"total"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got.Status != "confirmed" || got.Subtotal != 1000 || got.Discount != 100 || got.Total != 900 {
		t.Fatalf("unexpected order: %+v", got)
	}
}

func TestConfirmEmptyOrderBadRequest(t *testing.T) {
	t.Parallel()

	h := newHandler()
	body := `{"order_id":"o1","customer_id":"c1","lines":[]}`
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/orders", strings.NewReader(body))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("POST /orders status = %d, want 201", rec.Code)
	}

	req = httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/orders/o1/confirm", nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("confirm empty order status = %d, want 400; body=%s", rec.Code, rec.Body.String())
	}
}

func TestAddLineAfterConfirmConflict(t *testing.T) {
	t.Parallel()

	h := newHandler()
	body := `{"order_id":"o1","customer_id":"c1","lines":[{"product_id":"p1","name":"Widget","unit_price":100,"currency":"USD","quantity":1}]}`
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/orders", strings.NewReader(body))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("POST /orders status = %d, want 201", rec.Code)
	}

	req = httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/orders/o1/confirm", nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("confirm status = %d, want 200", rec.Code)
	}

	req = httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/orders/o1/lines", strings.NewReader(`{"product_id":"p2","name":"Widget","unit_price":100,"currency":"USD","quantity":1}`))
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusConflict {
		t.Fatalf("add after confirm status = %d, want 409; body=%s", rec.Code, rec.Body.String())
	}
}

func TestAddLineInvalidQuantityBadRequest(t *testing.T) {
	t.Parallel()

	h := newHandler()
	body := `{"order_id":"o1","customer_id":"c1","lines":[{"product_id":"p1","name":"Widget","unit_price":100,"currency":"USD","quantity":1}]}`
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/orders", strings.NewReader(body))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("POST /orders status = %d, want 201", rec.Code)
	}

	req = httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/orders/o1/lines", strings.NewReader(`{"product_id":"p2","name":"Widget","unit_price":100,"currency":"USD","quantity":0}`))
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("add invalid line status = %d, want 400; body=%s", rec.Code, rec.Body.String())
	}
}

func TestGetOrderMissing(t *testing.T) {
	t.Parallel()

	h := newHandler()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/orders/missing", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}
