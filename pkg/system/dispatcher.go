package system

import (
	// "fmt"
	"github.com/gradecak/watchdog/pkg/enforcer"
	"github.com/gradecak/watchdog/pkg/events"
	"github.com/sirupsen/logrus"
)

type EventDispatcher struct {
	enforcers map[int][]enforcer.Enforcer
}

func NewEventDispatcher(enforcers map[int][]enforcer.Enforcer) *EventDispatcher {
	return &EventDispatcher{enforcers: enforcers}
}

func (h *EventDispatcher) Register(e events.Event, enf enforcer.Enforcer) error {
	if enforcers, ok := h.enforcers[e.Type()]; ok {
		h.enforcers[e.Type()] = append(enforcers, enf)
		return nil
	} else {
		h.enforcers[e.Type()] = []enforcer.Enforcer{enf}
	}
	return nil
}

func (h EventDispatcher) Dispatch(e events.Event) {
	logrus.Info("Event of Type %v received", e.Type())
	if enforcers, ok := h.enforcers[e.Type()]; ok {
		for _, enforcer := range enforcers {
			enforcer.Enforce(e)
		}
		return
	}
	logrus.Warnf("Dispatcher recieved event not handled by any Enforcers: %v", e.Type())
}
