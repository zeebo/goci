package frontend

import (
	"github.com/zeebo/goci/app/httputil"
	"net/http"
)

//index shows the main homepage of goci
func index(w http.ResponseWriter, req *http.Request, ctx httputil.Context) (e *httputil.Error) {
	w.Header().Set("Content-Type", "text/html")
	if err := T("index/index.html").Execute(w, nil); err != nil {
		e = httputil.Errorf(err, "error executing index template")
	}
	return
}
