package worker

import "launchpad.net/mgo"

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
		All(&ws)
	return
}

func WorkImportPaths(ctx *Context) (ps []string, err error) {
	err = ctx.db.C(worklog).
		Find(nil).
		Distinct("importpath", &ps)
	return
}
