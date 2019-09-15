package api

import (
	"github.com/gradecak/watchdog/pkg/dispatcher/queue"
	"github.com/gradecak/watchdog/pkg/events"
)

type Dispatcher struct {
	dipatchQueue queue.DispatchQueue
}

func NewDispatcherAPI(q queue.DispatchQueue) *Dispatcher {
	return &Dispatcher{q}
}

func (d *Dispatcher) Queue(e *events.Event) {
	d.dipatchQueue.Add(e)
}
