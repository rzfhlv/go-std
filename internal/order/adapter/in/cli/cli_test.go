package cli_test

import (
	"context"
	"encoding/csv"
	"errors"
	"strings"
	"testing"

	"github.com/rzfhlv/go-std/internal/order/adapter/in/cli"
	"github.com/rzfhlv/go-std/internal/order/app"
)

// stubOrders is a fake app.OrderUseCases returning a fixed order list.
type stubOrders struct {
	orders []app.OrderOutput
	err    error
}

func (s *stubOrders) PlaceOrder(context.Context, app.PlaceOrderInput) (app.OrderOutput, error) {
	return app.OrderOutput{}, errors.New("not implemented")
}

func (s *stubOrders) AddLine(context.Context, string, app.LineInput) (app.OrderOutput, error) {
	return app.OrderOutput{}, errors.New("not implemented")
}

func (s *stubOrders) ConfirmOrder(context.Context, string) (app.OrderOutput, error) {
	return app.OrderOutput{}, errors.New("not implemented")
}

func (s *stubOrders) CancelOrder(context.Context, string, string) (app.OrderOutput, error) {
	return app.OrderOutput{}, errors.New("not implemented")
}

func (s *stubOrders) GetOrder(context.Context, string) (app.OrderOutput, error) {
	return app.OrderOutput{}, errors.New("not implemented")
}

func (s *stubOrders) ListOrders(context.Context) ([]app.OrderOutput, error) {
	return s.orders, s.err
}

func TestListEmptyMarker(t *testing.T) {
	t.Parallel()

	stub := &stubOrders{orders: []app.OrderOutput{
		{ID: "empty", CustomerID: "c1", Status: "pending", Currency: "USD"},
		{
			ID: "full", CustomerID: "c2", Status: "pending", Currency: "USD",
			Lines: []app.LineOutput{{ProductID: "p1", Name: "Widget", Quantity: 1, UnitPrice: 100, Subtotal: 100}},
		},
	}}

	var out strings.Builder
	if err := cli.New(stub, &out).Run(context.Background(), []string{"list"}); err != nil {
		t.Fatalf("Run error = %v", err)
	}

	rows := strings.Split(strings.TrimRight(out.String(), "\n"), "\n")
	if len(rows) != 3 {
		t.Fatalf("rows = %d, want 3 (header + 2); output=%q", len(rows), out.String())
	}

	header := strings.Fields(rows[0])
	if header[len(header)-1] != "EMPTY" {
		t.Fatalf("header last column = %q, want EMPTY; header=%q", header[len(header)-1], rows[0])
	}

	for _, tc := range []struct {
		rowIndex int
		want     string
	}{
		{rowIndex: 1, want: "yes"},
		{rowIndex: 2, want: "no"},
	} {
		fields := strings.Fields(rows[tc.rowIndex])
		got := fields[len(fields)-1]
		if got != tc.want {
			t.Errorf("row %d last column = %q, want %q; row=%q", tc.rowIndex, got, tc.want, rows[tc.rowIndex])
		}
	}
}

func TestExportCLI(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		stub     *stubOrders
		args     []string
		wantErr  bool
		wantRows int // including the header row
	}{
		{
			name: "multiple orders",
			stub: &stubOrders{orders: []app.OrderOutput{
				{ID: "o1", CustomerID: "c1", Status: "pending", Subtotal: 1000, Discount: 100, Total: 900, Currency: "USD"},
				{ID: "o2", CustomerID: "c2", Status: "confirmed", Subtotal: 500, Discount: 0, Total: 500, Currency: "USD"},
			}},
			args:     []string{"export"},
			wantRows: 3,
		},
		{
			name:     "no orders yields header only",
			stub:     &stubOrders{},
			args:     []string{"export"},
			wantRows: 1,
		},
		{
			name:     "extra args rejected",
			stub:     &stubOrders{},
			args:     []string{"export", "o1"},
			wantErr:  true,
			wantRows: 0,
		},
		{
			name:     "list error propagates",
			stub:     &stubOrders{err: errors.New("boom")},
			args:     []string{"export"},
			wantErr:  true,
			wantRows: 0,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var out strings.Builder
			err := cli.New(tc.stub, &out).Run(context.Background(), tc.args)
			if tc.wantErr {
				if err == nil {
					t.Fatal("Run error = nil, want error")
				}
				if out.Len() != 0 {
					t.Fatalf("output = %q, want empty on error", out.String())
				}
				return
			}
			if err != nil {
				t.Fatalf("Run error = %v", err)
			}
			if out.Len() == 0 {
				t.Fatal("no output written")
			}

			rows, err := csv.NewReader(strings.NewReader(out.String())).ReadAll()
			if err != nil {
				t.Fatalf("output is not valid CSV: %v; output=%q", err, out.String())
			}
			if len(rows) != tc.wantRows {
				t.Fatalf("csv rows = %d, want %d; output=%q", len(rows), tc.wantRows, out.String())
			}
			if len(rows) == 0 {
				t.Fatal("no csv rows to assert on")
			}
			wantHeader := []string{"id", "customer", "status", "subtotal", "discount", "total", "currency"}
			for i, want := range wantHeader {
				if rows[0][i] != want {
					t.Fatalf("header[%d] = %q, want %q", i, rows[0][i], want)
				}
			}
		})
	}
}

func TestExportCSVEscaping(t *testing.T) {
	t.Parallel()

	var out strings.Builder
	stub := &stubOrders{orders: []app.OrderOutput{
		{ID: "o,1", CustomerID: `c"1`, Status: "pending", Subtotal: 1, Total: 1, Currency: "USD"},
	}}
	if err := cli.New(stub, &out).Run(context.Background(), []string{"export"}); err != nil {
		t.Fatalf("Run error = %v", err)
	}

	rows, err := csv.NewReader(strings.NewReader(out.String())).ReadAll()
	if err != nil {
		t.Fatalf("output is not valid CSV: %v; output=%q", err, out.String())
	}
	if len(rows) != 2 {
		t.Fatalf("csv rows = %d, want 2", len(rows))
	}
	if rows[1][0] != "o,1" || rows[1][1] != `c"1` {
		t.Fatalf("escaped values not round-tripped: id=%q customer=%q", rows[1][0], rows[1][1])
	}
}
