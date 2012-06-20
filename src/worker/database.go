package worker

func GetRecentWork(ctx *Context, limit int) (ws []*Work, err error) {
	err = ctx.db.C(worklog).Find(nil).Sort(d{"$natural": -1}).Limit(10).All(&ws)
	return
}

func GetWorkFromBuild(ctx *Context, id string) (wk *Work, err error) {
	err = ctx.db.C(worklog).Find(d{"builds._id": id}).One(&wk)
	return
}

func CurrentWork(ctx *Context) (w []*mongoWork, err error) {
	err = ctx.db.C(workqueue).Find(nil).Limit(queue_size).All(&w)
	return
}
