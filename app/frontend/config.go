package frontend

import (
	"html/template"
	"net/http"
	"os"
	"path/filepath"
)

//Config stores the configuration for the frontend
var Config struct {
	Templates string //the path to the templates
	Static    string //the path to static assets
}

//tmap stores the mapping of template names to templates.
var tmap map[string]*template.Template

//Compile looks for templates and compiles them into the template map.
func Compile() (h http.Handler, err error) {
	//clean and make sure templates directory has a trailing slash
	Config.Templates = filepath.Clean(Config.Templates)
	if Config.Templates[len(Config.Templates)-1] != filepath.Separator {
		Config.Templates += string(filepath.Separator)
	}

	//add the templates into the map
	tmap = map[string]*template.Template{}
	filepath.Walk(Config.Templates, addTemplate)

	//add the static path to our handler
	fs := http.FileServer(http.Dir(Config.Static))
	mux.Handle("/static/", http.StripPrefix("/static", fs))

	//done!
	return mux, nil
}

func addTemplate(path string, info os.FileInfo, e error) error {
	//if we got any errors walking, just return it
	if e != nil {
		return e
	}

	//only look at html files not named _base.html
	base := filepath.Base(path)
	if base == "_base.html" || filepath.Ext(base) != ".html" || info.IsDir() {
		return nil
	}

	//parse the template file in with the base template
	baseTemplate := filepath.Join(Config.Templates, "_base.html")
	t, err := template.New("_base.html").ParseFiles(baseTemplate, path)
	if err != nil {
		return err
	}

	//store the template and strip of prefix directory
	name := path[len(Config.Templates):]
	tmap[name] = t

	return nil
}

//T looks up the template at the given name and returns it. It panics if the
//template can not be found.
func T(name string) *template.Template {
	if t, ok := tmap[name]; ok {
		return t.Lookup("_base.html")
	}
	panic("no template named: " + name)
}
