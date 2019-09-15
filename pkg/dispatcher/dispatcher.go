package dispatcher

import (
	"github.com/gradecak/watchdog/pkg/api"
	"github.com/gradecak/watchdog/pkg/dispatcher/queue"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"time"
)

var (
	EnforceTime = prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Namespace: "dispatcher",
		Subsystem: "enforcment",
		Name:      "time",
		Help:      "",
		Objectives: map[float64]float64{
			0.25: 0.0001,
			0.5:  0.0001,
			0.9:  0.0001,
			1:    0.0001,
		},
	}, []string{"event"})
)

func init() {
	prometheus.MustRegister(EnforceTime)
}

type Dispatcher struct {
	// TODO make DispatchQueue into a persistant storage
	q    queue.DispatchQueue
	pAPI *api.Policy
}

func NewDispatcher(d queue.DispatchQueue, p *api.Policy) *Dispatcher {
	return &Dispatcher{d, p}
}

func (d Dispatcher) Run() {
	for {
		event := d.q.Get()
		go func() {
			start := time.Now()
			err, enfs := d.pAPI.GetEnforcers(event.Prefix)
			if err != nil {
				logrus.Warnf("No enforcers registered for prefix %s", event.Prefix)
				return
			}
			// run the enforcer
			for _, enf := range enfs {
				_, err := enf(event)
				if err != nil {
					logrus.Error(err)
				}
			}
			EnforceTime.WithLabelValues(event.Prefix).Observe(float64(time.Since(start)))
		}()
	}
}
