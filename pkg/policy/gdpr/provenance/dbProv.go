package provenance

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gradecak/fission-workflows/pkg/provenance/graph"
	"github.com/huandu/go-sqlbuilder"
	"github.com/sirupsen/logrus"
)

const (
	executed_tbl = "executed"
	workflow_tbl = "workflows"
	children_tbl = "children"
	nodes_tbl    = "nodes"
)

type DbProv struct {
	*sql.DB
}

type DbConf struct {
	Init bool
	User string
	Pass string
	Db   string
	URL  string
}

type dbNode struct {
	WfId   int64
	TaskId int64
	Type   int64
	Op     int64
	Meta   string
	FnName string
	Task   string
}

func (n *dbNode) toGraphNode() *graph.Node {
	return &graph.Node{
		Type:   graph.Node_Type(n.Type),
		Op:     graph.Node_Op(n.Op),
		Meta:   n.Meta,
		FnName: n.FnName,
		Task:   n.Task,
	}
}

func NewDBProv(cnf *DbConf) (*DbProv, error) {
	connUrl := fmt.Sprintf("%s:%s@tcp(%s)/%s", cnf.User, cnf.Pass, cnf.URL, cnf.Db)
	logrus.Infof("Connecting to %s", connUrl)
	db, err := sql.Open("mysql", connUrl)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1000)
	err = db.Ping()
	if err != nil {
		err = initDb(cnf)
		if err != nil {
			return nil, err
		}
	}
	prov := &DbProv{db}
	if cnf.Init {
		logrus.Info("initiallising tables...")
		err := prov.initTables()
		if err != nil {
			return nil, err
		}
	}

	return prov, nil
}

func initDb(cnf *DbConf) error {
	connUrl := fmt.Sprintf("%s:%s@tcp(%s)/", cnf.User, cnf.Pass, cnf.URL)
	c, err := sql.Open("mysql", connUrl)
	if err != nil {
		return err
	}
	_, err = c.Exec("CREATE DATABASE watchdog;")
	if err != nil {
		return err
	}
	c.Close()
	return nil
}

func (db *DbProv) Executed(id string) map[int64][]*graph.Node {
	sb := sqlbuilder.NewSelectBuilder()
	sb.From(fmt.Sprintf("%s %s", executed_tbl, "e")).
		Select("n.*, w.wfID").
		Join(fmt.Sprintf("%s %s", workflow_tbl, "w"), "e.wfID = w.wfID").
		Join(fmt.Sprintf("%s %s", nodes_tbl, "n"), "w.taskID = n.taskID").
		Where("e.cID = " + sb.Var(id))
	sql, args := sb.Build()
	r, err := db.Query(sql, args...)
	defer r.Close()
	if err != nil {
		logrus.Error(err.Error())
		//empty response
		return make(map[int64][]*graph.Node)
	}

	executedWorkflows := map[int64][]*graph.Node{}
	for r.Next() {
		var n dbNode
		if err := r.Scan(&n.TaskId, &n.Type, &n.Meta, &n.FnName, &n.Task, &n.WfId); err != nil {
			logrus.Error(err.Error())
		}
		executedWorkflows[n.WfId] = append(executedWorkflows[n.WfId], n.toGraphNode())
	}
	return executedWorkflows
}

func (db *DbProv) GetWfTasks(id int64) []*graph.Node {
	sb := sqlbuilder.NewSelectBuilder()
	sb.From(fmt.Sprintf("%s %s", workflow_tbl, "w")).
		Select("n.type, n.meta, n.fnName, n.task").
		Join(tableAlias(nodes_tbl, "n"), "w.taskID = n.taskID").
		Where("w.wfID = " + sb.Var(id))
	sql, args := sb.Build()
	r, err := db.Query(sql, args...)
	defer r.Close()
	if err != nil {
		logrus.Error(err.Error())
		return []*graph.Node{}
	}

	wfTasks := []*graph.Node{}
	for r.Next() {
		var task graph.Node
		if err := r.Scan(&task.Type, &task.Meta, &task.FnName, &task.Task); err != nil {
			logrus.Error(err.Error())
		}
		wfTasks = append(wfTasks, &task)
	}
	return wfTasks
}

func (db *DbProv) GetWorkflowChildren(id int64) []int64 {
	sb := sqlbuilder.NewSelectBuilder()
	sb.From(tableAlias(children_tbl, "c")).
		Select("c.childID").
		Where("c.wfID = " + sb.Var(id))
	sql, args := sb.Build()
	r, err := db.Query(sql, args...)
	if err != nil {
		logrus.Error(err.Error())
		return []int64{}
	}
	defer r.Close()
	children := []int64{}
	for r.Next() {
		var i int64
		if err := r.Scan(i); err != nil {
			logrus.Error(err.Error())
		}
		children = append(children, i)
	}
	return children
}

// TODO add check to see if nodes/workflows already exist in table before inserting
func (db *DbProv) Merge(g *graph.Provenance) error {
	logrus.Infof("Merging Prov Event... %+v", g)
	ctx := context.TODO()
	conn, err := db.Conn(ctx)
	if err != nil {
		logrus.Error(err)
		return err
	}
	defer conn.Close()
	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		logrus.Info("WHAT THE FUCK")
		logrus.Error(err.Error())
		return err
	}
	logrus.Info("AAAAAAAAAAAAaaa")

	nodes := g.GetNodes()
	for entity, wfId := range g.GetExecuted() {
		logrus.Info("Has Executed")
		insertTasks := false
		insertWF := false
		wfIB := sqlbuilder.NewInsertBuilder().
			InsertInto(workflow_tbl).
			Cols("wfID", "taskID")
		taskIB := sqlbuilder.NewInsertBuilder().
			InsertInto(nodes_tbl).
			Cols("taskID", "type", "meta", "fnName", "task")
		for _, taskID := range g.GetWorkflowTaskIds(wfId) {
			logrus.Info("Has tasks")
			// check if workflow to task relationship already
			// exists, if not insert
			if !db.existsWfTask(wfId, taskID) {
				//relate task to workflow
				insertWF = true
				wfIB.Values(wfId, taskID)
			}

			//check if task node entry already exists, if not insert
			if !db.existsTask(taskID) {
				//insert nodes
				insertTasks = true
				task := nodes[taskID]
				taskIB.Values(taskID, task.GetType(), task.GetMeta(), task.GetFnName(), task.GetTask())
			}
		}

		// insert workflow
		if insertWF {
			sql, args := wfIB.Build()
			_, err := tx.Exec(sql, args...)
			if err != nil {
				return err
			}
		}
		// insert task nodes
		if insertTasks {
			// insert tasks
			sql, args := taskIB.Build()
			_, err = tx.Exec(sql, args...)
			if err != nil {
				return err
			}
		}
		// insert entity workflow relationship
		if !db.existsEntityRelationship(entity, wfId) {
			// link enitty to workflow
			sql, args := sqlbuilder.NewInsertBuilder().
				InsertInto(executed_tbl).
				Cols("cID", "wfID").
				Values(entity, wfId).
				Build()
			if _, err = tx.Exec(sql, args...); err != nil {
				return err
			}
		}

		// add workflow predecessors we reverse the predecessor
		// relationship and store predecessors as children for more
		// efficient policy enforcement
		if predecessors := g.GetWorkflowPredecessors(wfId); len(predecessors) > 0 {
			ib := sqlbuilder.NewInsertBuilder().
				InsertInto(children_tbl).
				Cols("wfID", "childID")
			for _, predecessorID := range predecessors {
				ib.Values(predecessorID, wfId)
			}
			sql, args := ib.Build()
			if _, err = tx.Exec(sql, args...); err != nil {
				return err
			}
		}

	}
	logrus.Info("RETURNING")
	err = tx.Commit()
	if err != nil {
		logrus.Error(err.Error())
	}
	logrus.Info("FUUUUCK")
	return err
}

func (db *DbProv) existsTask(taskID int64) bool {
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select("taskID").
		From(nodes_tbl).
		Where("taskID =" + sb.Var(taskID))
	query, args := sb.Build()
	return db.exists(query, args...)
}

func (db *DbProv) existsWfTask(wfId, taskId int64) bool {
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select("wfID").
		From(workflow_tbl).
		Where("wfID ="+sb.Var(wfId), "taskID ="+sb.Var(taskId))
	query, args := sb.Build()
	return db.exists(query, args...)
}

func (db *DbProv) existsEntityRelationship(entity string, wfId int64) bool {
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select("cID").
		From(executed_tbl).
		Where("cID="+sb.Var(entity), "wfID="+sb.Var(wfId))
	query, args := sb.Build()
	return db.exists(query, args...)
}

func (db *DbProv) exists(query string, args ...interface{}) bool {
	var r []byte
	ctx := context.TODO()
	conn, err := db.Conn(ctx)
	defer conn.Close()
	if err != nil {
		logrus.Error(err.Error())
	}
	err = conn.QueryRowContext(ctx, query, args...).Scan(&r)
	if err != nil {
		if err != sql.ErrNoRows {
			logrus.Error(err.Error())
		}
		return false
	}
	return true
}

func (db *DbProv) initTables() error {
	// executed table
	ctb := sqlbuilder.NewCreateTableBuilder()
	ctb.CreateTable(executed_tbl).
		IfNotExists().
		Define("cID", "VARCHAR(255)").
		Define("wfID", "BIGINT").
		Define("PRIMARY KEY (cID, wfID)")
	logrus.Info(ctb.String())
	_, err := db.Exec(ctb.String())
	if err != nil {
		return err
	}
	// workflows table
	ctb = sqlbuilder.NewCreateTableBuilder()
	ctb.CreateTable(workflow_tbl).
		IfNotExists().
		Define("wfID", "BIGINT").
		Define("taskID", "BIGINT").
		Define("PRIMARY KEY (wfID, taskID)")
	_, err = db.Exec(ctb.String())
	if err != nil {
		return err
	}
	// children table
	ctb = sqlbuilder.NewCreateTableBuilder()
	ctb.CreateTable(children_tbl).
		IfNotExists().
		Define("wfID", "BIGINT").
		Define("childID", "BIGINT").
		Define("PRIMARY KEY (wfID, childID)")
	_, err = db.Exec(ctb.String())
	if err != nil {
		return err
	}
	//node table
	ctb = sqlbuilder.NewCreateTableBuilder()
	ctb.CreateTable(nodes_tbl).
		IfNotExists().
		Define("taskID", "BIGINT", "PRIMARY KEY").
		Define("type", "BIGINT").
		Define("meta", "TEXT").
		Define("fnName", "VARCHAR(255)").
		Define("task", "VARCHAR(255)")
	_, err = db.Exec(ctb.String())
	if err != nil {
		return err
	}

	return nil
}

func tableAlias(tableName, alias string) string {
	return fmt.Sprintf("%s %s", tableName, alias)
}
