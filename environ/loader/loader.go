//package loader provides loading environment variables from a file
package loader

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

//Load loads a file with environment variables and sets the environment
//to be what is in the file.
func Load(filename string) (err error) {
	//open up the file
	f, err := os.Open(filename)
	if err != nil {
		return
	}
	defer f.Close()

	//clear out the old environment
	os.Clearenv()

	//read the lines of the file and set the environ
	buf := bufio.NewReader(f)
	for {
		line, err := buf.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		pos := strings.Index(line, "=")
		if pos == -1 {
			return fmt.Errorf("Invalid line: %q", line)
		}

		//set the environment variable stripping the newline character
		os.Setenv(line[:pos], line[pos+1:len(line)-1])
	}

	return
}
