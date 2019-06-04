package events

import (
	// "fmt"
	"github.com/fission/fission-workflows/pkg/types"
	"github.com/golang/protobuf/proto"
	"testing"
)

func TestParse(t *testing.T) {
	consentMsg := &types.ConsentMessage{
		ID: "Gopher",
		Status: &types.ConsentStatus{
			Status: types.ConsentStatus_Status(1),
		},
	}

	buf, err := proto.Marshal(consentMsg)
	if err != nil {
		t.Errorf("%v", err)
	}

	event, err := Parse(buf)
	if err != nil {
		t.Errorf("%v", err)
	}
	e := event.(*ConsentEvent)

	if consentMsg.ID != e.ID && consentMsg.Status != e.Status {
		t.Errorf("Parse Failed")
	}
}
