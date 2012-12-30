package frontend

import (
	"github.com/zeebo/goci/app/entities"
	"github.com/zeebo/goci/app/httputil"
	"html/template"
	"net/http"
	"net/http/httptest"
	"testing"
)

func makeGETRequest(path string) *http.Request {
	req, err := http.NewRequest("GET", path, nil)
	if err != nil {
		panic(err)
	}
	return req
}

type testQueryManager struct{}

func (testQueryManager) Index() ([]entities.TestResult, error)               { return nil, nil }
func (testQueryManager) SpecificWork(string) (*entities.Work, error)         { return nil, nil }
func (testQueryManager) Work(skip, limit int) ([]entities.WorkResult, error) { return nil, nil }
func (testQueryManager) Packages() (pkgListJobResult, error)                 { return nil, nil }

func init() {
	//stub out contextfunc for tests
	httputil.Config.ContextFunc = func(*http.Request) (c httputil.Context) { return }

	//stub out the template func
	T = func(unused string) *template.Template {
		return template.Must(template.New("").Parse("{{.}}"))
	}

	//stub out the query manager
	newManager = func(httputil.Context) queryManager {
		return testQueryManager{}
	}
}

func TestIndex(t *testing.T) {
	rec := httptest.NewRecorder()
	Mux.ServeHTTP(rec, makeGETRequest("/"))
	if rec.Code != 200 {
		t.Fatal("Invalid response code:", rec.Code)
	}
}

func TestSpecificWork(t *testing.T) {
	rec := httptest.NewRecorder()
	Mux.ServeHTTP(rec, makeGETRequest("/work/50dfac94346bea11bb000001"))
	if rec.Code != 200 {
		t.Fatal("Invalid response code:", rec.Code)
	}
}

func TestWork(t *testing.T) {
	rec := httptest.NewRecorder()
	Mux.ServeHTTP(rec, makeGETRequest("/work"))
	if rec.Code != 200 {
		t.Fatal("Invalid response code:", rec.Code)
	}
}

func TestSpecificImportResult(t *testing.T) {
	rec := httptest.NewRecorder()
	Mux.ServeHTTP(rec, makeGETRequest("/result/github.com/zeebo/irc@foo"))
	if rec.Code != 200 {
		t.Fatal("Invalid response code:", rec.Code)
	}
}

func TestImportResult(t *testing.T) {
	rec := httptest.NewRecorder()
	Mux.ServeHTTP(rec, makeGETRequest("/result/github.com/zeebo/irc"))
	if rec.Code != 200 {
		t.Fatal("Invalid response code:", rec.Code)
	}
}

func TestImage(t *testing.T) {
	rec := httptest.NewRecorder()
	Mux.ServeHTTP(rec, makeGETRequest("/image/github.com/zeebo/irc"))
	if rec.Code != 200 {
		t.Fatal("Invalid response code:", rec.Code)
	}
}

func TestHow(t *testing.T) {
	rec := httptest.NewRecorder()
	Mux.ServeHTTP(rec, makeGETRequest("/how"))
	if rec.Code != 200 {
		t.Fatal("Invalid response code:", rec.Code)
	}
}

func TestPkg(t *testing.T) {
	rec := httptest.NewRecorder()
	Mux.ServeHTTP(rec, makeGETRequest("/pkg"))
	if rec.Code != 200 {
		t.Fatal("Invalid response code:", rec.Code)
	}
}

func TestNotFound(t *testing.T) {
	paths := []string{"/doop"}
	for _, path := range paths {
		rec := httptest.NewRecorder()
		Mux.ServeHTTP(rec, makeGETRequest(path))
		if rec.Code != 404 {
			t.Error(path, "Invalid response code:", rec.Code)
		}
	}
}
