package gdpr

import (
	"errors"
	// "fmt"
	"github.com/golang/protobuf/proto"
	"github.com/gradecak/fission-workflows/pkg/provenance/graph"
	"github.com/gradecak/fission-workflows/pkg/types"
	"github.com/gradecak/watchdog/pkg/events"
	"github.com/gradecak/watchdog/pkg/policy"
	"github.com/gradecak/watchdog/pkg/policy/gdpr/provenance"
	"github.com/sirupsen/logrus"
	// "gopkg.in/yaml.v2"
	// "net/http"
	// "strings"
)

type Policy struct {
	prov *provenance.MemProv
}

func NewPolicy(provAPI *provenance.MemProv) policy.Policy {
	return Policy{provAPI}
}

func (p Policy) Actions(e string) (error, []policy.Enforcer) {
	switch e {
	case "CONSENT":
		return nil, []policy.Enforcer{p.consentEnforcer()}
	case "PROVENANCE":
		return nil, []policy.Enforcer{p.provenanceEnforcer()}
	}
	return events.ERR_UNKNOWN, nil
}

//
//Consent Enforcer
//

type Meta struct {
	Revoke string `yaml:"revoke"`
}

func (p Policy) consentEnforcer() policy.Enforcer {
	return func(e *events.Event) ([]*policy.Result, error) {
		// parse event payload
		conEvent := &types.ConsentMessage{}
		err := proto.Unmarshal(e.Payload, conEvent)
		if err != nil {
			return nil, errors.New("Received event is not Consent Event")
		}
		// only take action on revoked status
		if status := conEvent.Status.Status; status == types.ConsentStatus_REVOKED {
			logrus.Info("Revoked status")
			writeTasks := []*graph.Node{}

			// b := strings.Split(conEvent.ID, "/$/")
			// id, startTime := b[0], b[1]
			// filter tasks
			for _, tasks := range p.prov.Executed(conEvent.ID) {
				// for _, tasks := range p.prov.Executed(id) {
				logrus.Info("Num tasks for ID %s -- %s", conEvent.ID, len(tasks))
				for _, task := range tasks {
					if task.GetOp() == graph.Node_WRITE {
						// m := &Meta{}
						logrus.Infof("Task Meta %v", task.GetMeta())
						// err := yaml.Unmarshal([]byte(task.GetMeta()), m)
						// if err != nil {
						// 	logrus.Error(err.Error())
						// 	return nil, err
						// }
						// resp, err := http.Get(fmt.Sprintf("%s/%s", task.GetMeta(), startTime))
						if err != nil {
							logrus.Error(err.Error())
						}
						// resp.Body.Close()
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
	return func(e *events.Event) ([]*policy.Result, error) {
		pg := &graph.Provenance{}
		err := proto.Unmarshal(e.Payload, pg)
		if err != nil {
			return nil, errors.New("Received event could not be cast to ProvEvent")
		}
		err = p.prov.Merge(pg)
		if err != nil {
			logrus.Errorf("Merge Error; reason: %v", err.Error())
		}
		return []*policy.Result{}, err

	}
}
