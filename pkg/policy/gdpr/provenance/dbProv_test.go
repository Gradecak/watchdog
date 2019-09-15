package provenance

import (
	"github.com/gradecak/fission-workflows/pkg/provenance/graph"
	"github.com/sirupsen/logrus"
	"testing"
)

func TestConnect(t *testing.T) {
	db, err := NewDBProv()
	if err != nil {
		t.Error(err)
	}
	// db.Executed("poo")
	logrus.Info(db.GetWfTasks(1))
}

func TestWrite(t *testing.T) {
	db, err := NewDBProv()
	if err != nil {
		t.Error(err)
	}

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
