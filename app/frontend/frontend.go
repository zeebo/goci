//package frontend is the web frontend for goci
package frontend

import (
	"github.com/zeebo/goci/app/httputil"
	"net/http"
)

//Mux is the ServeMux for the frontend
var Mux = http.NewServeMux()

//register all the handlers with the serve mux
func init() {
	fs := http.FileServer(Config)

	Mux.Handle("/", httputil.Handler(index))
	Mux.Handle("/static/", http.StripPrefix("/static", fs))
}
