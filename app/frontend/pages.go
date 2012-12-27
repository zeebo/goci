package frontend

import (
	"github.com/zeebo/goci/app/entities"
	"github.com/zeebo/goci/app/httputil"
	"labix.org/v2/mgo"
	"net/http"
	"net/url"
	"sort"
	"time"
)

type (
	d map[string]interface{}
	l []interface{}
)

//grab loads a key from the form and returns the result. panics if the result
//is not present or if there are multiple entries. it should only be used for
//grabbing values out of parsed arguments from the pat package.
func grab(parsed url.Values, key string) (val string) {
	if vals := parsed[":"+key]; len(vals) == 1 {
		val = vals[0]
		return
	}
	panic("too many or few values for key")
}

//not found displays a pretty 404 page
func notFound(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	http.NotFound(w, req)
}

//index shows the main homepage of goci
func index(w http.ResponseWriter, req *http.Request, ctx httputil.Context) (e *httputil.Error) {
	if req.URL.Path != "/" {
		notFound(w, req)
		return
	}

	var res []entities.TestResult
	err := ctx.DB.C("TestResult").Find(nil).Sort("-when").Limit(20).All(&res)
	if err != nil {
		e = httputil.Errorf(err, "couldn't query for test results")
	}

	w.Header().Set("Content-Type", "text/html")
	if err := T("index/index.html").Execute(w, res); err != nil {
		e = httputil.Errorf(err, "error executing index template")
	}
	return
}

//work shows recent work items
func work(w http.ResponseWriter, req *http.Request, ctx httputil.Context) (e *httputil.Error) {
	w.Header().Set("Content-Type", "text/html")
	if err := T("index/index.html").Execute(w, nil); err != nil {
		e = httputil.Errorf(err, "error executing index template")
	}
	return
}

//specific_work shows a work item with the given key
func specifc_work(w http.ResponseWriter, req *http.Request, ctx httputil.Context) (e *httputil.Error) {
	w.Header().Set("Content-Type", "text/html")
	if err := req.ParseForm(); err != nil {
		e = httputil.Errorf(err, "error parsing form")
		return
	}

	key := grab(req.Form, "key")
	if err := T("index/index.html").Execute(w, key); err != nil {
		e = httputil.Errorf(err, "error executing index template")
	}
	return
}

//result shows recent result items
func result(w http.ResponseWriter, req *http.Request, ctx httputil.Context) (e *httputil.Error) {
	w.Header().Set("Content-Type", "text/html")

	if err := T("index/index.html").Execute(w, nil); err != nil {
		e = httputil.Errorf(err, "error executing index template")
	}
	return
}

//import_result shows recent result items for an import path
func import_result(w http.ResponseWriter, req *http.Request, ctx httputil.Context) (e *httputil.Error) {
	w.Header().Set("Content-Type", "text/html")
	if err := req.ParseForm(); err != nil {
		e = httputil.Errorf(err, "error parsing form")
		return
	}

	imp := grab(req.Form, "import")
	if err := T("index/index.html").Execute(w, imp); err != nil {
		e = httputil.Errorf(err, "error executing index template")
	}
	return
}

//specific_import_result shows a result item for an import path and given revision
func specific_import_result(w http.ResponseWriter, req *http.Request, ctx httputil.Context) (e *httputil.Error) {
	w.Header().Set("Content-Type", "text/html")
	if err := req.ParseForm(); err != nil {
		e = httputil.Errorf(err, "error parsing form")
		return
	}

	imp, rev := grab(req.Form, "import"), grab(req.Form, "rev")
	if err := T("index/index.html").Execute(w, struct{ Imp, Rev string }{imp, rev}); err != nil {
		e = httputil.Errorf(err, "error executing index template")
	}
	return
}

//image returns an image representing the most recent build status for an import path
func image(w http.ResponseWriter, req *http.Request, ctx httputil.Context) (e *httputil.Error) {
	w.Header().Set("Content-Type", "text/html")
	if err := req.ParseForm(); err != nil {
		e = httputil.Errorf(err, "error parsing request form")
		return
	}

	imp := grab(req.Form, "import")
	if err := T("index/index.html").Execute(w, imp); err != nil {
		e = httputil.Errorf(err, "error executing index template")
	}
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

//pkg displays a list of import paths tested by goci
func pkg(w http.ResponseWriter, req *http.Request, ctx httputil.Context) (e *httputil.Error) {
	w.Header().Set("Content-Type", "text/html")

	var res pkgListJobResult
	_, err := ctx.DB.C("TestResult").Find(nil).MapReduce(pkgListJob, &res)
	if err != nil {
		e = httputil.Errorf(err, "error grabbing package list")
		return
	}
	sort.Sort(res)
	if err := T("pkg/pkg.html").Execute(w, res); err != nil {
		e = httputil.Errorf(err, "error executing index template")
	}
	return
}

//how displays a page showing how to use the service
func how(w http.ResponseWriter, req *http.Request, ctx httputil.Context) (e *httputil.Error) {
	w.Header().Set("Content-Type", "text/html")

	if err := T("index/index.html").Execute(w, nil); err != nil {
		e = httputil.Errorf(err, "error executing index template")
	}
	return
}
