package nats

import (
	// "fmt"
	"github.com/golang/protobuf/proto"
	"github.com/gradecak/fission-workflows/pkg/types"
	stan "github.com/nats-io/stan.go"
	"testing"
)

func TestConnect(t *testing.T) {
	consentMsg := &types.ConsentMessage{
		ID: "poo",
		Status: &types.ConsentStatus{
			Status: types.ConsentStatus_Status(1),
		},
	}

	buf, err := proto.Marshal(consentMsg)
	if err != nil {
		t.Errorf("%v", err)
	}
	conn, err := stan.Connect("test-cluster", "tester", stan.NatsURL("127.0.0.1"))
	if err != nil {
		t.Errorf("%v", err)
	}

	conn.Publish("CONSENT", buf)

}
