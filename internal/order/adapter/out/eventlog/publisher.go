// Package eventlog is an outbound adapter that writes order domain events to an
// io.Writer as a simple stand-in for a real message broker.
package eventlog

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/rzfhlv/go-std/internal/order/domain"
	"github.com/rzfhlv/go-std/internal/order/port"
)

// Publisher writes each domain event as a single line to an io.Writer.
type Publisher struct {
	out io.Writer
}

// New returns a Publisher writing to out.
func New(out io.Writer) *Publisher { return &Publisher{out: out} }

var _ port.EventPublisher = (*Publisher)(nil)

// Publish implements port.EventPublisher.
func (p *Publisher) Publish(_ context.Context, events []domain.Event) error {
	for _, event := range events {
		if _, err := fmt.Fprintf(p.out, "event %s at %s\n", event.EventName(), event.OccurredAt().Format(time.RFC3339)); err != nil {
			return err
		}
	}
	return nil
}
