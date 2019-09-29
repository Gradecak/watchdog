package provenance

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gradecak/fission-workflows/pkg/provenance/graph"
	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"
)

type Node struct {
	// gorm.Model
	Id                 int64 `gorm:"primary_key;not_auto_increment"`
	Op                 int32
	Meta               string
	Type               int32
	FnName             string
	Task               string
	WorkflowTasks      []*Node `gorm:"many2many:wf_tasks;association_jointable_foreignkey:task_id"`
	WorkflowPredecssor []*Node `gorm:"many2many:wf_predecessor;association_jointable_foreignkey:predecssor_id"`
}

func (n *Node) toGraphNode() *graph.Node {
	return &graph.Node{
		Type:   graph.Node_Type(n.Type),
		Op:     graph.Node_Op(n.Op),
		Meta:   n.Meta,
		FnName: n.FnName,
		Task:   n.Task,
	}
}

type User struct {
	// gorm.Model
	ConsentId string  `gorm:"primary_key"`
	Workflows []*Node `gorm:"many2many:executed"`
	// Workflows []Workflow `gorm:"many2many:executed"`
}

type DbProv struct {
	*gorm.DB
}

// import (
// 	"context"
// 	"database/sql"
// 	"fmt"
// 	_ "github.com/go-sql-driver/mysql"
// 	"github.com/gradecak/fission-workflows/pkg/provenance/graph"
// 	"github.com/huandu/go-sqlbuilder"
// 	"github.com/sirupsen/logrus"
// )

type DbConf struct {
	Init bool
	User string
	Pass string
	Db   string
	URL  string
}

func NewDBProv(cnf *DbConf) (*DbProv, error) {
	connUrl := fmt.Sprintf("%s:%s@tcp(%s)/%s", cnf.User, cnf.Pass, cnf.URL, cnf.Db)
	db, err := gorm.Open("mysql", connUrl)
	db.DB().SetMaxIdleConns(10)
	db.DB().SetMaxOpenConns(100)
	db.LogMode(false)
	if err != nil {
		return nil, err
	}

	if cnf.Init {
		db.AutoMigrate(&User{}, &Node{})
	}
	return &DbProv{db}, nil
}

func (db *DbProv) Executed(id string) map[int64][]*graph.Node {
	user := &User{}
	res := db.Where(&User{ConsentId: id}).
		Preload("Workflows").
		Preload("Workflows.WorkflowTasks").
		Find(user)
	if res.Error != nil {
		logrus.Error(res.Error)
		return nil
	}

	executed := make(map[int64][]*graph.Node)
	for _, workflow := range user.Workflows {
		for _, task := range workflow.WorkflowTasks {
			executed[workflow.Id] = append(executed[workflow.Id], task.toGraphNode())
		}
	}
	return executed
}

func (db *DbProv) GetWfTasks(id int64) []*graph.Node {
	return nil
}

func (db *DbProv) GetWorkflowChildren(id int64) []int64 {
	return nil
}

func (db *DbProv) Merge(g *graph.Provenance) error {
	var (
		users     = []*User{}
		nodes     = g.GetNodes()
		workflows = []*Node{}
	)

	for entity, wfId := range g.GetExecuted() {
		tasks := []*Node{}
		for _, taskID := range g.GetWorkflowTaskIds(wfId) {
			taskNode := nodes[taskID]
			// append our tasks to list of tasks
			tasks = append(tasks, &Node{
				Id:     taskID,
				Op:     int32(taskNode.GetOp()),
				Type:   int32(taskNode.GetType()),
				Meta:   taskNode.Meta,
				FnName: taskNode.FnName,
				Task:   taskNode.Task,
			})
		}

		wfNode := nodes[wfId]
		workflows = append(workflows, &Node{
			Id:            wfId,
			Type:          int32(wfNode.GetType()),
			WorkflowTasks: tasks,
		})

		// if predecessors := g.GetWor
		users = append(users, &User{
			ConsentId: entity,
			Workflows: workflows,
		})
	}

	for _, u := range users {
		for retry := 0; retry < 2; retry++ {
			res := db.Save(u)
			if res.Error == nil {
				if retry == 2 {
					panic(res.Error)
				}
				break
			}
		}

	}
	return nil
}

// func initDb(cnf *DbConf) error {
// 	connUrl := fmt.Sprintf("%s:%s@tcp(%s)/", cnf.User, cnf.Pass, cnf.URL)
// 	c, err := sql.Open("mysql", connUrl)
// 	if err != nil {
// 		return err
// 	}
// 	_, err = c.Exec("CREATE DATABASE watchdog;")
// 	if err != nil {
// 		return err
// 	}
// 	c.Close()
// 	return nil
// }

// func (db *DbProv) Executed(id string) map[int64][]*graph.Node {
// 	sb := sqlbuilder.NewSelectBuilder()
// 	sb.From(fmt.Sprintf("%s %s", executed_tbl, "e")).
// 		Select("n.*, w.wfID").
// 		Join(fmt.Sprintf("%s %s", workflow_tbl, "w"), "e.wfID = w.wfID").
// 		Join(fmt.Sprintf("%s %s", nodes_tbl, "n"), "w.taskID = n.taskID").
// 		Where("e.cID = " + sb.Var(id))
// 	sql, args := sb.Build()
// 	r, err := db.Query(sql, args...)
// 	defer r.Close()
// 	if err != nil {
// 		logrus.Error(err.Error())
// 		//empty response
// 		return make(map[int64][]*graph.Node)
// 	}

// 	executedWorkflows := map[int64][]*graph.Node{}
// 	for r.Next() {
// 		var n dbNode
// 		if err := r.Scan(&n.TaskId, &n.Type, &n.Meta, &n.FnName, &n.Task, &n.WfId); err != nil {
// 			logrus.Error(err.Error())
// 		}
// 		executedWorkflows[n.WfId] = append(executedWorkflows[n.WfId], n.toGraphNode())
// 	}
// 	return executedWorkflows
// }

// func (db *DbProv) GetWfTasks(id int64) []*graph.Node {
// 	sb := sqlbuilder.NewSelectBuilder()
// 	sb.From(fmt.Sprintf("%s %s", workflow_tbl, "w")).
// 		Select("n.type, n.meta, n.fnName, n.task").
// 		Join(tableAlias(nodes_tbl, "n"), "w.taskID = n.taskID").
// 		Where("w.wfID = " + sb.Var(id))
// 	sql, args := sb.Build()
// 	r, err := db.Query(sql, args...)
// 	defer r.Close()
// 	if err != nil {
// 		logrus.Error(err.Error())
// 		return []*graph.Node{}
// 	}

// 	wfTasks := []*graph.Node{}
// 	for r.Next() {
// 		var task graph.Node
// 		if err := r.Scan(&task.Type, &task.Meta, &task.FnName, &task.Task); err != nil {
// 			logrus.Error(err.Error())
// 		}
// 		wfTasks = append(wfTasks, &task)
// 	}
// 	return wfTasks
// }

// func (db *DbProv) GetWorkflowChildren(id int64) []int64 {
// 	sb := sqlbuilder.NewSelectBuilder()
// 	sb.From(tableAlias(children_tbl, "c")).
// 		Select("c.childID").
// 		Where("c.wfID = " + sb.Var(id))
// 	sql, args := sb.Build()
// 	r, err := db.Query(sql, args...)
// 	if err != nil {
// 		logrus.Error(err.Error())
// 		return []int64{}
// 	}
// 	defer r.Close()
// 	children := []int64{}
// 	for r.Next() {
// 		var i int64
// 		if err := r.Scan(i); err != nil {
// 			logrus.Error(err.Error())
// 		}
// 		children = append(children, i)
// 	}
// 	return children
// }
