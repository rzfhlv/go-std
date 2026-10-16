// Package systemclock provides the system-clock implementation of the order
// module's clock outbound port.
package systemclock

import "time"

// Clock reports the current system time.
type Clock struct{}

// New returns a system Clock.
func New() Clock { return Clock{} }

// Now implements port.Clock.
func (Clock) Now() time.Time { return time.Now() }
