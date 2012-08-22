package httputil_test

import (
	"fmt"
	"github.com/zeebo/goci/app/httputil"
)

// http://goci.me/some/path
func ExampleAbsolute() {
	httputil.Config.Domain = "goci.me"
	fmt.Println(httputil.Absolute("/some/path"))
}
