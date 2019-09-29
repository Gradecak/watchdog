package dispatcher

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"sync"
)

var (
	CurrentEnfrocers = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "dispatcher",
		Subsystem: "concurrent",
		Name:      "enforcers",
		Help:      "",
	})
)

func init() {
	prometheus.MustRegister(CurrentEnfrocers)
}

type Monitor struct {
	active    int
	activeMu  *sync.RWMutex
	maxActive int
	waitingMu *sync.Mutex
	waiting   []chan bool
	unblocked chan bool
	done      chan bool
}

func NewMonitor(max int) *Monitor {
	return &Monitor{
		active:    0,
		activeMu:  &sync.RWMutex{},
		maxActive: max,
		waitingMu: &sync.Mutex{},
		waiting:   []chan bool{},
		done:      make(chan bool, 1000),
	}
}

func (m *Monitor) Run() {
	for {
		select {
		case <-m.done:
			// block until an enforcer calls done
			m.activeMu.RLock()
			m.waitingMu.Lock()
			if m.active < m.maxActive-400 {
				for _, waiting := range m.waiting {
					logrus.Info("NOTIFYING SLEEPERS")
					waiting <- true
				}
				m.waiting = []chan bool{}
			}
			m.waitingMu.Unlock()
			m.activeMu.RUnlock()

		}
	}
}

func (m *Monitor) Permitted(wakeup chan bool) bool {
	// m.activeMu.RLock()
	if m.active < m.maxActive {
		// m.activeMu.RUnlock()
		return true
	}
	// m.activeMu.RUnlock()
	// add the new channel to our list of clients waiting for unblock
	m.waitingMu.Lock()
	m.waiting = append(m.waiting, wakeup)
	m.waitingMu.Unlock()
	return false
}

func (m *Monitor) AddEnforcer() {
	m.activeMu.Lock()
	m.active++
	m.activeMu.Unlock()
	CurrentEnfrocers.Inc()
}

func (m *Monitor) RemoveEnforcer() {
	m.activeMu.Lock()
	m.active--
	m.done <- true
	m.activeMu.Unlock()
	CurrentEnfrocers.Dec()
}
