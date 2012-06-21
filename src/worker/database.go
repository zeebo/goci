package worker

import (
	"launchpad.net/mgo"
	"time"
)

func GetRecentWork(ctx *Context, limit int) (ws []*Work, err error) {
	err = ctx.db.C(worklog).
		Find(nil).
		Sort(d{"$natural": -1}).
		Limit(limit).
		All(&ws)
	return
}

func GetWorkFromBuild(ctx *Context, id string) (wk *Work, err error) {
	err = ctx.db.C(worklog).
		Find(d{"builds._id": id}).
		One(&wk)
	return
}

func CurrentWork(ctx *Context) (w []*mongoWork, err error) {
	err = ctx.db.C(workqueue).
		Find(nil).
		Limit(queue_size).
		All(&w)
	return
}

func CountWork(ctx *Context) (n int, err error) {
	n, err = ctx.db.C(worklog).
		Find(nil).
		Count()
	return
}

func WorkInRange(ctx *Context, low, hi int) (ws []*Work, err error) {
	count := hi - low + 1
	err = ctx.db.C(worklog).
		Find(nil).
		Skip(low).
		Limit(count).
		Sort(d{"when": -1}).
		All(&ws)
	return
}

func workWithImportPathSel(ctx *Context, path string) (q *mgo.Query) {
	q = ctx.db.C(worklog).
		Find(d{"importpath": path})
	return
}

func CountWorkFor(ctx *Context, path string) (n int, err error) {
	q := workWithImportPathSel(ctx, path)
	n, err = q.Count()
	return
}

func WorkWithImportPath(ctx *Context, path string) (ws []*Work, err error) {
	q := workWithImportPathSel(ctx, path)
	err = q.All(&ws)
	return
}

func WorkWithImportPathInRange(ctx *Context, path string, low, hi int) (ws []*Work, err error) {
	q := workWithImportPathSel(ctx, path)
	count := hi - low + 1
	err = q.
		Skip(low).
		Limit(count).
		Sort(d{"when": -1}).
		All(&ws)
	return
}

type WorkStatusResult struct {
	ImportPath string
	When       time.Time
	Status     WorkStatus
}

var work_status_job = mgo.MapReduce{
	Map: `function() { emit(this.importpath, {
			when: this.when,
			status: this.status
		});
	}`,
	Reduce: `function(key, values) {
		var result = values.shift();
		values.forEach(function(value) {
			if (result.when < value.when) {
				result = value;
			}
		});
		return result;
	}`,
}

func WorkStatusList(ctx *Context) (ws []WorkStatusResult, err error) {
	var res []struct {
		ImportPath string `bson:"_id"`
		Value      struct {
			When   time.Time
			Status WorkStatus
		}
	}

	_, err = ctx.db.C(worklog).
		Find(nil).
		MapReduce(work_status_job, &res)

	if err != nil {
		return
	}

	//copy in
	for _, v := range res {
		ws = append(ws, WorkStatusResult{
			ImportPath: v.ImportPath,
			When:       v.Value.When,
			Status:     v.Value.Status,
		})
	}

	return
}
