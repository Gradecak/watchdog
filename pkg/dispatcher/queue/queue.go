package queue

import (
	"github.com/gradecak/watchdog/pkg/events"
)

type DispatchQueue interface {
	Add(*events.Event)
	Get() *events.Event
}
