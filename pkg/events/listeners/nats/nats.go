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
	Prefixes []string
}

type EventListener struct {
	conn     stan.Conn
	prefixes []string
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
	return &EventListener{conn: conn, prefixes: cfg.Prefixes}, nil
}

func (el *EventListener) Listen(e chan *events.Event) {
	//On event recieved callback
	eventCB := func(prefix string, e chan *events.Event) func(*stan.Msg) {
		return func(m *stan.Msg) {
			e <- &events.Event{prefix, m.Data}
		}
	}

	// start event stream subscriptions
	for _, prefix := range el.prefixes {
		el.conn.Subscribe(prefix, eventCB(prefix, e))
	}
}
