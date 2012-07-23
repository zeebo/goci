package builder

import "net/url"

//baseUrl is the url we use for downloading binaries and tarballs
var baseUrl *url.URL

//parse our baseUrl in from the environment
func init() {
	var err error
	baseUrl, err = url.Parse(env("BASE_URL", "http://builder.goci.me"))
	if err != nil {
		bail(err)
	}
}

//urlWithPath makes a copy of the baseUrl and sets the path to the provided path
//and returns the string representation
func urlWithPath(path string) string {
	urlCopy := *baseUrl
	urlCopy.Path = path
	return urlCopy.String()
}
