package builder

import (
	"encoding/json"
	"github.com/zeebo/goci/app/rpc"
	"path/filepath"
	"strings"
)

//loadConfig grabs the Config data for the given import path by walking up the
//directory tree of the import path and overwriting older data with newer data.
func (b Builder) loadConfig(baseImport, subImport string) (c rpc.Config, err error) {
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
func loadConfigAt(c *rpc.Config, dir string) (err error) {
	file := filepath.Join(dir, ".goci")

	//check if the file exists first (we won't have a race for it being deleted)
	if !World.Exists(file) {
		return
	}

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
