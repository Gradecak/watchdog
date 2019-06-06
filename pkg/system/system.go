package system

import (
	"github.com/gradecak/watchdog/pkg/events"
	"github.com/gradecak/watchdog/pkg/policy"
	"github.com/sirupsen/logrus"
)

func (s System) dispatchEnforcer(e events.Event) {
	logrus.Infof("Event of Type %v received", e.Type())
	for _, enforcer := range s.policy.Actions(e.Type()) {
		enforcer(e)
	}
	// logrus.Warnf("Dispatcher recieved event not handled by any Enforcers: %v", e.Type())
}

type System struct {
	policy   policy.Policy
	incoming []events.Listener
}

func New(policy policy.Policy, listeners []events.Listener) *System {
	return &System{policy, listeners}
}

func (s System) Run() error {
	// subscribe to event sourcers
	events := make(chan events.Event)
	for _, listener := range s.incoming {
		listener.Listen(events)
	}

	//dispatch enforcers for processing events
	for {
		select {
		case event := <-events:
			go s.dispatchEnforcer(event)
		}

	}
}
