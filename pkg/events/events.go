package events

import (
	"errors"
	"github.com/fission/fission-workflows/pkg/provenance/graph"
	"github.com/fission/fission-workflows/pkg/types"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/sirupsen/logrus"
)

const (
	Event_CONSENT    = 0
	Event_PROVENANCE = 1
)

var (
	ERR_PARSE = errors.New("Could not decode incoming message to an Event")
)

type Event interface {
	Type() int
}

type EventParser interface {
	Parse(data []byte) (Event, error)
}

func parse(data []byte) (proto.Message, error) {
	i := &any.Any{}
	err := proto.Unmarshal(data, i)
	if err != nil {
		return nil, err
	}
	logrus.Infof(i.GetTypeUrl())
	return i, nil
}

//
// Consent Event
//
type ConsentEvent struct {
	Msg *types.ConsentMessage
}

func (e ConsentEvent) Type() int {
	return Event_CONSENT
}

func (e ConsentEvent) Parse(data []byte) (Event, error) {
	cm := &types.ConsentMessage{}
	err := proto.Unmarshal(data, cm)
	if err != nil {
		return nil, err
	}
	return &ConsentEvent{cm}, nil
}

//
// Provenance Event
//
type ProvEvent struct {
	Msg *graph.Provenance
}

func (e ProvEvent) Type() int {
	return Event_PROVENANCE
}

func (e ProvEvent) Parse(data []byte) (Event, error) {
	p := &graph.Provenance{}
	err := proto.Unmarshal(data, p)
	if err != nil {
		logrus.Errorf("ERROR : %v", err)
		return nil, err
	}
	return &ProvEvent{p}, nil
}
