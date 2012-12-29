//package frontend is the web frontend for goci
package frontend

import (
	"code.google.com/p/gorilla/pat"
	"github.com/zeebo/goci/app/httputil"
	"net/http"
)

//Mux is the handler for the frontend
var Mux = pat.New()

//register all the handlers with the serve mux
func init() {
	Mux.Add("GET", "/static/", http.StripPrefix("/static", http.FileServer(Config)))
	Mux.Add("GET", "/work/{key:.+}", httputil.Handler(specificWork))
	Mux.Add("GET", "/work", httputil.Handler(work))
	Mux.Add("GET", "/result/{import:[^@]+}@{rev:.*}", httputil.Handler(specificImportResult))
	Mux.Add("GET", "/result/{import:[^@]+}", httputil.Handler(importResult))
	Mux.Add("GET", "/result", httputil.Handler(result))
	Mux.Add("GET", "/image/{import:.+}", httputil.Handler(image))
	Mux.Add("GET", "/how", httputil.Handler(how))
	Mux.Add("GET", "/pkg", httputil.Handler(pkg))
	Mux.Add("GET", "/", httputil.Handler(index))
}
