package worker

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
