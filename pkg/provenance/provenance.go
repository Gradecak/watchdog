package memprov

import (
	"github.com/fission/fission-workflows/pkg/provenance/graph"
	"github.com/sirupsen/logrus"
	"sync"
)

type MemProv struct {
	executed     map[string][]int64    // map entities to workflow executions
	workflows    map[int64][]int64     // map workflow to tasks
	nodes        map[int64]*graph.Node // map taskid to task node
	predecessors map[int64][]int64     // map a workflow to its predecessor
	mux          *sync.Mutex
}

func NewProv() *MemProv {
	return &MemProv{
		executed:     make(map[string][]int64),
		workflows:    make(map[int64][]int64),
		nodes:        make(map[int64]*graph.Node),
		predecessors: make(map[int64][]int64),
		mux:          &sync.Mutex{},
	}
}

func (m *MemProv) Executed(id string) map[int64][]*graph.Node {
	wfs := map[int64][]*graph.Node{}
	for _, wfID := range m.executed[id] {
		for _, taskID := range m.workflows[wfID] {
			wfs[wfID] = append(wfs[wfID], m.nodes[taskID])
		}
	}
	return wfs
}

func (m *MemProv) GetWfTasks(id int64) []*graph.Node {
	tasks := []*graph.Node{}
	for _, taskID := range m.workflows[id] {
		tasks = append(tasks, m.nodes[taskID])
	}
	return tasks
}

func (m *MemProv) Merge(g *graph.Provenance) error {
	m.mux.Lock()
	for entity, wfId := range g.GetExecuted() {
		m.addExecuted(entity, wfId)
		// add tasks
		for _, taskID := range g.GetWorkflowTaskIds(wfId) {
			if _, ok := m.nodes[taskID]; !ok {
				m.nodes[taskID] = g.GetNodes()[taskID]
				m.workflows[wfId] = append(m.workflows[wfId], taskID)
			}
		}
		// add predecessor relationship
		for _, predID := range g.GetWorkflowPredecessors(wfId) {
			if !m.hasPredecessor(wfId, predID) {
				m.predecessors[wfId] = append(m.predecessors[wfId], predID)
			}
		}
	}
	m.mux.Unlock()
	logrus.Infof("%+v\n", m)
	return nil
}

func (m *MemProv) hasPredecessor(wfId int64, pred int64) bool {
	if predecessors, ok := m.predecessors[wfId]; ok {
		for _, p := range predecessors {
			if p == pred {
				return true
			}
		}
	}
	return false
}

func (m *MemProv) addExecuted(entity string, wfID int64) {
	for _, executedID := range m.executed[entity] {
		if executedID == wfID {
			return
		}
	}
	m.executed[entity] = append(m.executed[entity], wfID)
}
