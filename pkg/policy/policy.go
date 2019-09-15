package policy

import (
	"github.com/gradecak/watchdog/pkg/events"
	"time"
)

type Result struct {
	Payload interface{}
	Ts      time.Time
}

type Enforcer func(e *events.Event) ([]*Result, error)

type Policy interface {
	Actions(string) (error, []Enforcer) // event type to list of enforcers
}
