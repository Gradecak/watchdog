package policy

import (
	"github.com/gradecak/watchdog/pkg/events"
	"time"
)

// type Policy map[int][]enforcer.Enforcer

type Violation struct {
	Msg string
	Ts  time.Time
}

type Enforcer func(events.Event) ([]*Violation, error)

type Policy interface {
	Actions(int) []Enforcer // event type to list of enforcers
}
