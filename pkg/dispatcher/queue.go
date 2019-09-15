package dispatcher

import (
	"github.com/gradecak/watchdog/pkg/events"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	NumQueued = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "dispatcher",
		Subsystem: "queue",
		Name:      "lenght",
		Help:      "",
	})
)

func init() {
	prometheus.MustRegister(NumQueued)
}

//TODO add persistant storage for queue items

type DispatchQueue chan *events.Event

func NewDispatchQueue(size int) DispatchQueue {
	q := make(chan *events.Event, size)
	return q
}

func (d DispatchQueue) Add(e *events.Event) {
	defer NumQueued.Inc()
	d <- e
}

func (d DispatchQueue) Get() *events.Event {
	defer NumQueued.Dec()
	return <-d
}
