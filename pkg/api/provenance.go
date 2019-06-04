package api

import (
	"github.com/fission/fission-workflows/pkg/provenance/graph"
)

type ProvenanceStore interface {
	// for a given entity find all of the Workflows that the entity was
	// involved in.  returns map[WorkflowId]Tasks
	Executed(string) map[int64]*graph.Node
	GetWfTasks(int64) []*graph.Node
	GetWfPredecessors(int64) []int64
	// us to extend the current provenance map state with the newly generate
	// provenance data
	Merge(*graph.Provenance) error
}
