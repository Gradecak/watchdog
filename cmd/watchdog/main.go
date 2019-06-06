package main

import (
	"github.com/gradecak/watchdog/pkg/events"
	"github.com/gradecak/watchdog/pkg/events/listeners/nats"
	"github.com/gradecak/watchdog/pkg/policy/gdpr"
	"github.com/gradecak/watchdog/pkg/provenance"
	"github.com/gradecak/watchdog/pkg/system"
)

func main() {
	natsConf := &nats.Config{
		Cluster: "test-cluster",
		Client:  "watchdog",
		URL:     "127.0.0.1",
		Matchers: map[string]events.EventParser{
			"CONSENT":    events.ConsentEvent{},
			"PROVENANCE": events.ProvEvent{},
		},
	}
	listener, err := nats.New(natsConf)
	if err != nil {
		panic(err.Error())
	}
	policy := gdpr.NewPolicy(memprov.NewProv())
	sys := system.New(policy, []events.Listener{listener})
	sys.Run()
}
