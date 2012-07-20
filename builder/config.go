package builder

import (
	"encoding/json"
	"path/filepath"
	"strings"
)

//Config represents a json struct that is read in by the Builder. It can describe
//things like notifcations or other configuration data for the test. It is loaded
//from files named `.goci` in the package directory. The file is loaded by descending
//the package heiarchy loading in the config and overwriting values as it goes,
//so that config files lower in the path inherit the values from higher in the
//path.
type Config struct {
	NotifyJabber string   `json:",omitempty"` // a jabber address for an XMPP message
	NotifyOn     []string `json:",omitempty"` // fail, pass, change, always
	NotifyURL    string   `json:",omitempty"` // a URL that will be posted with the result data
}

//loadConfig grabs the Config data for the given import path by walking up the
//directory tree of the import path and overwriting older data with newer data.
func (b Builder) loadConfig(baseImport, subImport string) (c Config, err error) {
	//we start at the source directory of the baseImport and work our way up
	//to the subImport
	base := filepath.Join(b.gopath, "src", baseImport)
	extra := subImport[len(baseImport):]

	//explode the extra import path into directories
	importParts := strings.Split(extra, string(filepath.Separator))

	//make a container for all of the parts
	parts := make([]string, 0, len(importParts)+1)

	//put the base in front, and the import path directories after
	parts = append(parts, base)
	parts = append(parts, importParts...)

	//loop over each starting from the lowest and iteratively apply the config
	for i := 0; i <= len(parts); i++ {
		at := filepath.Join(parts[:i]...)
		err = loadConfigAt(&c, at)
		if err != nil {
			return
		}
	}

	return
}

//loadConfigAt looks for a .goci file in the given directory and attemps to load
//its data into the passed in config.
func loadConfigAt(c *Config, dir string) (err error) {
	file := filepath.Join(dir, ".goci")
	//fi

	//don't report errors in loading the file
	r, err := World.Open(file)
	if err != nil {
		//the file should exist so report this error and ditch the build
		return
	}

	//decode the data and close the file
	err = json.NewDecoder(r).Decode(c)
	r.Close()
	return
}
