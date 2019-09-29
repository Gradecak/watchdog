package provenance

import (
	"github.com/gradecak/fission-workflows/pkg/provenance/graph"
	"github.com/sirupsen/logrus"
	"testing"
)

var (
	conn = &DbConf{
		Init: true,
		User: "root",
		Pass: "12345",
		Db:   "watchdog",
		URL:  "127.0.0.1:3306",
	}
	db = &DbProv{}
)

func TestConnect(t *testing.T) {
	var err error
	db, err = NewDBProv(conn)
	if err != nil {
		t.Error(err)
	}
	// db.Executed("poo")
	logrus.Info(db.GetWfTasks(1))
}

func TestWrite(t *testing.T) {
	task := &graph.Node{
		Type:   3,
		Meta:   "poo",
		FnName: "poo",
		Task:   "poo",
	}

	graph := &graph.Provenance{
		Nodes:    map[int64]*graph.Node{12: task},
		WfTasks:  map[int64]*graph.IDs{13: &graph.IDs{[]int64{12}}},
		Executed: map[string]int64{"radical": 13},
	}

	if err := db.Merge(graph); err != nil {
		t.Error(err)
	}
}

func TestUpdate(t *testing.T) {
	task := &graph.Node{
		Type:   3,
		Meta:   "shit",
		FnName: "shit",
		Task:   "shit",
	}

	graph := &graph.Provenance{
		Nodes:    map[int64]*graph.Node{14: task},
		WfTasks:  map[int64]*graph.IDs{15: &graph.IDs{[]int64{14}}},
		Executed: map[string]int64{"radical": 15},
	}

	if err := db.Merge(graph); err != nil {
		t.Error(err)
	}
}
