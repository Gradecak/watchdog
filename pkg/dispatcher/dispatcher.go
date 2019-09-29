package dispatcher

import (
	"github.com/gradecak/watchdog/pkg/api"
	"github.com/gradecak/watchdog/pkg/dispatcher/queue"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"time"
)

var (
	TotalTime = prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Namespace: "dispatcher",
		Subsystem: "total",
		Name:      "time",
		Help:      "",
		Objectives: map[float64]float64{
			0.25: 0.0001,
			0.5:  0.0001,
			0.9:  0.0001,
			1:    0.0001,
		},
	}, []string{"event"})
	EnforceTime = prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Namespace: "dispatcher",
		Subsystem: "enforce",
		Name:      "time",
		Help:      "",
		Objectives: map[float64]float64{
			0.25: 0.0001,
			0.5:  0.0001,
			0.9:  0.0001,
			1:    0.0001,
		},
	}, []string{"event"})
	EnforceCount = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "dispatcher",
		Subsystem: "enforcment",
		Name:      "count",
		Help:      "",
	}, []string{"event"})
)

func init() {
	prometheus.MustRegister(TotalTime, EnforceCount, EnforceTime)
}

type Dispatcher struct {
	// TODO make DispatchQueue into a persistant storage
	q       queue.DispatchQueue
	pAPI    *api.Policy
	monitor *Monitor
}

func NewDispatcher(d queue.DispatchQueue, p *api.Policy) *Dispatcher {
	monitor := NewMonitor(1000)
	return &Dispatcher{
		q:       d,
		pAPI:    p,
		monitor: monitor,
	}
}

func (d *Dispatcher) Run() {
	go d.monitor.Run()
	for {
		// if we are at capacity wait until a slot frees up
		wakeChan := make(chan bool)
		if !d.monitor.Permitted(wakeChan) {
			<-wakeChan
		}

		var (
			queueEvent = d.q.Get()
			event      = queueEvent.Event
		)

		d.monitor.AddEnforcer()
		go func(m *Monitor) {
			start := time.Now()

			// fetch enforcers from policy
			err, enfs := d.pAPI.GetEnforcers(event.Prefix)
			if err != nil {
				logrus.Warnf("No enforcers registered for prefix %s", event.Prefix)
				return
			}

			// process registered enforcers
			for _, enf := range enfs {
				// run the enforcer
				_, err := enf(event)
				if err != nil {
					logrus.Error(err)
				}
				//log metrics
				EnforceCount.WithLabelValues(event.Prefix).Inc()

			}
			m.RemoveEnforcer()
			TotalTime.WithLabelValues(event.Prefix).Observe(float64(time.Since(queueEvent.Queued)))
			EnforceTime.WithLabelValues(event.Prefix).Observe(float64(time.Since(start)))
		}(d.monitor)
	}
}
