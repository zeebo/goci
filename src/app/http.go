package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"thegoods.biz/httpbuf"
)

type runner interface {
	Run(http.ResponseWriter, *http.Request, *Context)
}

func perform(r runner, w http.ResponseWriter, req *http.Request) {
	ctx := init_context(req)
	defer ctx.Close()

	var buf httpbuf.Buffer
	defer buf.Apply(w)

	r.Run(&buf, req, ctx)
	ctx.Save(req, w)
}

type handlerFunc func(http.ResponseWriter, *http.Request, *Context)

func (h handlerFunc) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	perform(h, w, req)
}

func (h handlerFunc) Run(w http.ResponseWriter, req *http.Request, ctx *Context) {
	h(w, req, ctx)
}

type auth_func func(*http.Request, *Context) bool

type dispatcher struct {
	GET, POST, HEAD handlerFunc
	auth            auth_func
}

func (d *dispatcher) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	perform(d, w, req)
}

func (d *dispatcher) Run(w http.ResponseWriter, req *http.Request, ctx *Context) {
	if d.auth != nil && !d.auth(req, ctx) {
		perform_status(w, ctx, http.StatusNotFound)
		return
	}

	switch {
	case req.Method == "GET" && d.GET != nil:
		d.GET(w, req, ctx)
	case req.Method == "POST" && d.POST != nil:
		d.POST(w, req, ctx)
	case req.Method == "HEAD" && d.HEAD != nil:
		d.HEAD(w, req, ctx)
	default:
		perform_status(w, ctx, http.StatusNotFound)
	}
}

func wrap_runner(r runner) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		perform(r, w, req)
	})
}

func handleGet(pattern string, handler http.Handler, name string) {
	router.Add("GET", pattern, handler).Name(name)
}

func handlePost(pattern string, handler http.Handler, name string) {
	router.Add("POST", pattern, handler).Name(name)
}

func handleRequest(pattern string, handler http.Handler, name string) {
	handleGet(pattern, handler, name)
	handlePost(pattern, handler, name)
}

type staticHandler struct {
	h   http.Handler
	buf *httpbuf.Buffer
	o   sync.Once
}

func (s *staticHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if s.buf == nil {
		s.o.Do(func() {
			s.buf = new(httpbuf.Buffer)
			s.h.ServeHTTP(s.buf, req)

			//remove any Set-Cookie headers from the map as this
			//could cause session jacking!
			delete(s.buf.Header(), "Set-Cookie")
		})
	}
	s.buf.Apply(w)
}

//cache takes a handler and wraps it in a staticHandler which caches the output
//forever. Obviously this is unsuitable for any dynamic content.
func cache(h http.Handler) http.Handler {
	return &staticHandler{h: h}
}

func timelog(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		defer func(n time.Time) { log.Println("time spent:", time.Since(n)) }(time.Now())
		h.ServeHTTP(w, req)
	})
}

// Serves static files from filesystemDir when any request is made matching
// requestPrefix
func serve_static(requestPrefix, filesystemDir string) {
	fileServer := http.FileServer(http.Dir(filesystemDir))
	handler := http.StripPrefix(requestPrefix, fileServer)
	http.Handle(requestPrefix+"/", handler)
}

func serve_static_cached(requestPrefix, filesystemDir string) {
	fileServer := http.FileServer(http.Dir(filesystemDir))
	handler := http.StripPrefix(requestPrefix, fileServer)
	cached := path_cache(handler)
	http.Handle(requestPrefix+"/", cached)
}

type pathCacher struct {
	h http.Handler
	m map[string]*staticHandler
}

func (p *pathCacher) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	s, ex := p.m[req.URL.Path]
	if !ex {
		s = &staticHandler{h: p.h}
		p.m[req.URL.Path] = s
	}
	s.ServeHTTP(w, req)
}

//path_cache takes a handler and makes each path a separate cached handler.
func path_cache(h http.Handler) http.Handler {
	return &pathCacher{
		h: h,
		m: map[string]*staticHandler{},
	}
}

func reverse(name string, things ...interface{}) string {
	strs := make([]string, len(things))
	for i, th := range things {
		strs[i] = fmt.Sprint(th)
	}
	u, err := router.GetRoute(name).URL(strs...)
	if err != nil {
		panic(err)
	}
	return u.Path
}
