package system

import (
	"github.com/gradecak/watchdog/pkg/events"
)

type System struct {
	dispatcher *EventDispatcher
	incoming   events.Listener
	// Storage
}

func New(d *EventDispatcher, l events.Listener) *System {
	return &System{d, l}
}

func (s System) Run() error {
	events := make(chan events.Event)
	s.incoming.Listen(events)

	for {
		select {
		case event := <-events:
			go s.dispatcher.Dispatch(event)
		}

	}
}
