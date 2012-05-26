package env

import (
	"bufio"
	"flag"
	"io"
	"os"
	"strings"
)

var envfile = flag.String("env", "", "Path to env file")

func init() {
	flag.Parse()

	if *envfile == "" {
		return
	}

	os.Clearenv()
	if err := load_env(*envfile); err != nil {
		panic(err)
	}
}

func load_env(path string) (err error) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()
	b := bufio.NewReader(f)
	var line string

	for {
		line, err = b.ReadString('\n')
		if err == io.EOF {
			err = nil
			break
		}
		if err != nil {
			return
		}

		chunks := strings.SplitN(line, "=", 2) //split into 2 things
		os.Setenv(chunks[0], strings.TrimSpace(chunks[1]))
	}
	return
}
