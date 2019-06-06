package gdpr

import (
	"errors"
	"github.com/fission/fission-workflows/pkg/provenance/graph"
	"github.com/fission/fission-workflows/pkg/types"
	"github.com/gradecak/watchdog/pkg/api"
	"github.com/gradecak/watchdog/pkg/events"
	"github.com/gradecak/watchdog/pkg/policy"
	"github.com/sirupsen/logrus"
)

type Policy struct {
	prov api.ProvenanceStore
}

func NewPolicy(provAPI api.ProvenanceStore) policy.Policy {
	return Policy{provAPI}
}

func (p Policy) Actions(e int) []policy.Enforcer {
	switch e {
	case events.Event_CONSENT:
		return []policy.Enforcer{p.consentEnforcer()}
	case events.Event_PROVENANCE:
		return []policy.Enforcer{p.provenanceEnforcer()}
	}
	return nil
}

//
//Consent Enforcer
//
func (p Policy) consentEnforcer() policy.Enforcer {
	return func(e events.Event) ([]*policy.Violation, error) {
		// ensure event recieved is correct type
		conEvent, ok := e.(*events.ConsentEvent)
		if !ok {
			return nil, errors.New("Received event is not Consent Event")
		}

		// only take action on revoked status
		if status := conEvent.Msg.Status.Status; status == types.ConsentStatus_REVOKED {

			writeTasks := []*graph.Node{}
			// filter tasks
			for _, tasks := range p.prov.Executed(conEvent.Msg.ID) {
				for _, task := range tasks {
					if task.GetOp() == graph.Node_WRITE {
						writeTasks = append(writeTasks, task)
					}
				}
			}
		}
		return nil, nil
	}
}

//
// Provenance Consistency Enforcer
//
func (p Policy) provenanceEnforcer() policy.Enforcer {
	return func(e events.Event) ([]*policy.Violation, error) {
		logrus.Info("Processing provenance update event\n")
		if pg, ok := e.(*events.ProvEvent); ok {
			return []*policy.Violation{}, p.prov.Merge(pg.Msg)
		}
		err := errors.New("Received event could not be cast to ProvEvent")
		logrus.Warn(err.Error())
		return nil, err
	}
}
