package frontend

import (
	"github.com/zeebo/goci/app/entities"
	"github.com/zeebo/goci/app/httputil"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"sort"
	"time"
)

type pkgListJobResult []*struct {
	ImportPath string `bson:"_id"`
	Value      struct {
		When   time.Time
		Status string
	}
}

func (p pkgListJobResult) Len() int      { return len(p) }
func (p pkgListJobResult) Swap(i, j int) { p[i], p[j] = p[j], p[i] }
func (p pkgListJobResult) Less(i, j int) bool {
	return p[i].Value.When.After(p[j].Value.When)
}

var newManager = func(ctx httputil.Context) queryManager {
	return &mgoQueryManager{db: ctx.DB}
}

type queryManager interface {
	Index() ([]entities.TestResult, error)
	SpecificWork(id string) (*entities.Work, error)
	Work(skip, limit int) ([]entities.WorkResult, error)
	Packages() (pkgListJobResult, error)
}

type mgoQueryManager struct {
	db *mgo.Database
}

func (m *mgoQueryManager) Index() (res []entities.TestResult, err error) {
	err = m.db.C("TestResult").Find(nil).Sort("-when").Limit(20).All(&res)
	return
}

func (m *mgoQueryManager) SpecificWork(id string) (work *entities.Work, err error) {
	err = m.db.C("Work").FindId(bson.ObjectIdHex(id)).One(&work)
	return
}

func (m *mgoQueryManager) Work(skip, limit int) (res []entities.WorkResult, err error) {
	err = m.db.C("WorkResult").Find(nil).Sort("-when").Skip(skip).Limit(limit).All(&res)
	return
}

var pkgListJob = &mgo.MapReduce{
	Map: `function() { emit(this.importpath, {
			when: this.revdate,
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

func (m *mgoQueryManager) Packages() (res pkgListJobResult, err error) {
	_, err = m.db.C("TestResult").Find(nil).MapReduce(pkgListJob, &res)
	sort.Sort(res)
	return
}
