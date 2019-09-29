package dispatcher

import (
	"github.com/gradecak/watchdog/pkg/dispatcher/queue"
	"github.com/gradecak/watchdog/pkg/events"
	"github.com/prometheus/client_golang/prometheus"
	"time"
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

type DispatchQueue chan *queue.QueueEvent

func NewDispatchQueue(size int) DispatchQueue {
	q := make(chan *queue.QueueEvent, size)
	return q
}

func (d DispatchQueue) Add(e *events.Event) {
	defer NumQueued.Inc()
	d <- &queue.QueueEvent{e, time.Now()}
}

func (d DispatchQueue) Get() *queue.QueueEvent {
	defer NumQueued.Dec()
	return <-d
}
