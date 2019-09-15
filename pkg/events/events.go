package events

import (
	"errors"
	// "github.com/golang/protobuf/proto"
	// "github.com/golang/protobuf/ptypes/any"
	// "github.com/gradecak/fission-workflows/pkg/provenance/graph"
	// "github.com/gradecak/fission-workflows/pkg/types"
	// "github.com/sirupsen/logrus"
)

var (
	ERR_PARSE   = errors.New("Could not decode incoming message to an Event")
	ERR_UNKNOWN = errors.New("Recievied prefix not handled by policy")
)

type Event struct {
	Prefix  string // what event stream the event came from
	Payload []byte // payload recieved
}

// type EventParser interface {
// 	Parse(data []byte) (Event, error)
// }

// func (e ConsentEvent) Parse(data []byte) (Event, error) {
// 	cm := &types.ConsentMessage{}
// 	err := proto.Unmarshal(data, cm)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &ConsentEvent{cm}, nil
// }
