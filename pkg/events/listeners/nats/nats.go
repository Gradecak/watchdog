package nats

import (
	"github.com/gradecak/watchdog/pkg/events"
	stan "github.com/nats-io/stan.go"
	"github.com/sirupsen/logrus"
	"time"
)

const (
	defaultClient     = "fes"
	defaultCluster    = "fes-cluster"
	reconnectInterval = 10 * time.Second
)

type Config struct {
	Cluster string
	Client  string
	URL     string
	// list of prefixes that we
	// subscribe to for events
	Matchers map[string]events.EventParser
}

type EventListener struct {
	conn     stan.Conn
	matchers map[string]events.EventParser
}

func New(cfg *Config) (*EventListener, error) {
	logrus.Info("Connecting to NATS Event cluster...")
	conn, err := stan.Connect(cfg.Cluster, cfg.Client, stan.NatsURL(cfg.URL),
		stan.SetConnectionLostHandler(func(_ stan.Conn, reason error) {
			logrus.Fatalf("Connection lost, reason: %v", reason)
		}))

	if err != nil {
		return nil, err
	}
	return &EventListener{conn: conn, matchers: cfg.Matchers}, nil
}

func (el *EventListener) Listen(e chan events.Event) {
	//On event recieved callback
	eventCB := func(p events.EventParser, e chan events.Event) func(*stan.Msg) {
		return func(m *stan.Msg) {
			event, err := p.Parse(m.Data)
			if err != nil {
				return
			}
			e <- event
		}
	}

	// start event stream subscriptions
	for prefix, parser := range el.matchers {
		el.conn.Subscribe(prefix, eventCB(parser, e))
	}
}
