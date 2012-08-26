//package frontend is the web frontend for goci
package frontend

import (
	"net/http"
)

//Handler is the net/http Handler for the frontend. It expects the urls to have
//any prefix stripped before handling the request.
var Handler http.Handler = new(http.ServeMux)
