package aggregator

import (
	"github.com/gradecak/watchdog/pkg/api"
	"github.com/gradecak/watchdog/pkg/events"
	// "github.com/sirupsen/logrus"
)

type EventAggregator struct {
	dispatchAPI *api.Dispatcher
	incoming    []events.Listener
}

func NewEventAggregator(dispatchAPI *api.Dispatcher, listeners []events.Listener) *EventAggregator {
	return &EventAggregator{dispatchAPI, listeners}
}

func (s EventAggregator) Run() error {
	// subscribe to event sourcers
	events := make(chan *events.Event)
	for _, listener := range s.incoming {
		listener.Listen(events)
	}

	//dispatch enforcers for processing events
	for {
		select {
		case event := <-events:
			// logrus.Info("New Event. Queueing")
			s.dispatchAPI.Queue(event)
		}
	}
}
