package queue

import (
	"github.com/gradecak/watchdog/pkg/events"
	"time"
)

type DispatchQueue interface {
	Add(*events.Event)
	Get() *QueueEvent
}

type QueueEvent struct {
	Event  *events.Event
	Queued time.Time
}
