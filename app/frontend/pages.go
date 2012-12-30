package frontend

import (
	"github.com/zeebo/goci/app/httputil"
	"net/http"
	"net/url"
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
	m := newManager(ctx)

	res, err := m.Index()
	if err != nil {
		e = httputil.Errorf(err, "couldn't query for test results")
		return
	}

	w.Header().Set("Content-Type", "text/html")
	if err := T("index/index.html").Execute(w, res); err != nil {
		e = httputil.Errorf(err, "error executing index template")
	}
	return
}

//work shows recent work items
func work(w http.ResponseWriter, req *http.Request, ctx httputil.Context) (e *httputil.Error) {
	m := newManager(ctx)

	res, err := m.Work(0, 20)
	if err != nil {
		e = httputil.Errorf(err, "couldn't query for work results")
		return
	}

	w.Header().Set("Content-Type", "text/html")
	if err := T("work/work.html").Execute(w, res); err != nil {
		e = httputil.Errorf(err, "error executing index template")
	}
	return
}

//specificWork shows a work item with the given key
func specificWork(w http.ResponseWriter, req *http.Request, ctx httputil.Context) (e *httputil.Error) {
	w.Header().Set("Content-Type", "text/html")
	if err := req.ParseForm(); err != nil {
		e = httputil.Errorf(err, "error parsing form")
		return
	}
	m := newManager(ctx)

	work, err := m.SpecificWork(grab(req.Form, "key"))
	if err != nil {
		e = httputil.Errorf(err, "error grabbing work item")
		return
	}

	if err := T("work/specific_work.html").Execute(w, work); err != nil {
		e = httputil.Errorf(err, "error executing index template")
	}
	return
}

//result shows recent result items
func result(w http.ResponseWriter, req *http.Request, ctx httputil.Context) (e *httputil.Error) {
	w.Header().Set("Content-Type", "text/html")

	if err := T("result/result.html").Execute(w, nil); err != nil {
		e = httputil.Errorf(err, "error executing index template")
	}
	return
}

//importResult shows recent result items for an import path
func importResult(w http.ResponseWriter, req *http.Request, ctx httputil.Context) (e *httputil.Error) {
	w.Header().Set("Content-Type", "text/html")
	if err := req.ParseForm(); err != nil {
		e = httputil.Errorf(err, "error parsing form")
		return
	}

	imp := grab(req.Form, "import")
	if err := T("result/import_result.html").Execute(w, imp); err != nil {
		e = httputil.Errorf(err, "error executing index template")
	}
	return
}

//specificImportResult shows a result item for an import path and given revision
func specificImportResult(w http.ResponseWriter, req *http.Request, ctx httputil.Context) (e *httputil.Error) {
	w.Header().Set("Content-Type", "text/html")
	if err := req.ParseForm(); err != nil {
		e = httputil.Errorf(err, "error parsing form")
		return
	}

	imp, rev := grab(req.Form, "import"), grab(req.Form, "rev")
	if err := T("result/specific_import_result.html").Execute(w, []string{imp, rev}); err != nil {
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
	_ = imp
	return
}

//pkg displays a list of import paths tested by goci
func pkg(w http.ResponseWriter, req *http.Request, ctx httputil.Context) (e *httputil.Error) {
	m := newManager(ctx)

	res, err := m.Packages()
	if err != nil {
		e = httputil.Errorf(err, "error grabbing package list")
		return
	}

	w.Header().Set("Content-Type", "text/html")
	if err := T("pkg/pkg.html").Execute(w, res); err != nil {
		e = httputil.Errorf(err, "error executing index template")
	}
	return
}

//how displays a page showing how to use the service
func how(w http.ResponseWriter, req *http.Request, ctx httputil.Context) (e *httputil.Error) {
	w.Header().Set("Content-Type", "text/html")

	if err := T("how/how.html").Execute(w, nil); err != nil {
		e = httputil.Errorf(err, "error executing index template")
	}
	return
}
