package frontend

import (
	"html/template"
	"path/filepath"
)

//tmap caches the mapping of template names to templates.
var tmap = map[string]*template.Template{}

//T looks up the template at the given name and returns it. It panics if there
//are any errors parsing the template.
func T(name string) *template.Template {
	baseTemplate := filepath.Join(Config.Templates, "_base.html")
	path := filepath.Join(Config.Templates, name)

	//look up the template name in the cache if debug is not set
	if t, ok := tmap[name]; ok && !Config.Debug {
		return t.Lookup("_base.html")
	}

	//parse the template and add it to the cache
	t, err := template.New("_base.html").ParseFiles(baseTemplate, path)
	if err != nil {
		panic(err)
	}
	tmap[name] = t

	return t.Lookup("_base.html")
}
