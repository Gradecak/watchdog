package system

// import (
// 	"github.com/gradecak/watchdog/pkg/api"
// 	"github.com/gradecak/watchdog/pkg/events"
// 	//"github.com/sirupsen/logrus"
// )

// type System struct {
// 	dispatchAPI *api.Dispatcher
// 	incoming    []events.Listener
// }

// func NewSystem(dispatchAPI *api.Dispatcher, listeners []events.Listener) *System {
// 	return &System{dispatchAPI, listeners}
// }

// func (s System) Run() error {
// 	// subscribe to event sourcers
// 	events := make(chan *events.Event)
// 	for _, listener := range s.incoming {
// 		listener.Listen(events)
// 	}

// 	//dispatch enforcers for processing events
// 	for {
// 		select {
// 		case event := <-events:
// 			s.dispatchAPI.Queue(event)
// 		}
// 	}
// }
