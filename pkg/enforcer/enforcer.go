package enforcer

import (
	"github.com/gradecak/watchdog/pkg/events"
	"time"
)

type Enforcer interface {
	Enforce(events.Event) ([]*Violation, error)
}

// Violation Event. Notifies when a breach has happenedp
type Violation struct {
	Msg string
	Ts  time.Time
}

func NewViolation(m string) *Violation {
	return &Violation{
		Msg: m,
		Ts:  time.Now(),
	}
}
