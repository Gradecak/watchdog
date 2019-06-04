package policy

import (
	"errors"
	"github.com/gradecak/watchdog/pkg/api"
	"github.com/gradecak/watchdog/pkg/enforcer"
	"github.com/gradecak/watchdog/pkg/events"
	"github.com/sirupsen/logrus"
)

type Policy = map[int][]enforcer.Enforcer

func NewGDPRPolicy(provAPI api.Provenance) Policy {
	return Policy{
		events.Event_CONSENT:    []enforcer.Enforcer{NewConsentEnforcer(provAPI)},
		events.Event_PROVENANCE: []enforcer.Enforcer{NewProvenanceEnforcer(provAPI)},
	}
}

//
//Consent Enforcer
//
type ConsentEnforcer struct {
	prov api.Provenance
}

func NewConsentEnforcer(papi api.Provenance) *ConsentEnforcer {
	return &ConsentEnforcer{papi}
}

func (en ConsentEnforcer) Enforce(e events.Event) ([]*enforcer.Violation, error) {
	logrus.Infof("Received Event %+v", e)
	return nil, nil
}

//
// Provenance Consistency Enforcer
//
type ProvEnforcer struct {
	prov api.Provenance
}

func NewProvenanceEnforcer(papi api.Provenance) *ProvEnforcer {
	return &ProvEnforcer{papi}
}

func (p ProvEnforcer) Enforce(e events.Event) ([]*enforcer.Violation, error) {
	logrus.Info("Processing provenance update event\n")
	if pg, ok := e.(*events.ProvEvent); ok {
		return []*enforcer.Violation{}, p.prov.Merge(pg.Msg)
	}
	err := errors.New("Received event could not be cast to ProvEvent")
	logrus.Warn(err.Error())
	return nil, err
}
