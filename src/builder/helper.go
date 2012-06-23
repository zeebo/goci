package builder

import (
	"os"
	"sort"
)

func env(key, def string) string {
	if k := os.Getenv(key); k != "" {
		return k
	}
	return def
}

func unique(in []string) (out []string) {
	if len(in) == 0 {
		out = in
		return
	}
	sort.Strings(in)
	//do a seed
	i, p := 1, in[0]

	//loop over them and insert into the right spot
	for _, v := range in[1:] {
		if p != v {
			in[i] = v
			i, p = i+1, v
		}
	}

	out = in[:i]
	return
}
