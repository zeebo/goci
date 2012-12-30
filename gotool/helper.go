package gotool

import (
	"sort"
)

//unique sorts and uniques the slice in place.
func unique(in []string) (out []string) {
	if len(in) == 0 {
		out = in
		return
	}

	//sort the input
	sort.Strings(in)

	//do a seed
	i, p := 1, in[0]

	//loop over them keeping the unique ones
	for _, v := range in[1:] {
		if p != v {
			i, p, in[i] = i+1, v, v
		}
	}

	//reslice
	out = in[:i]
	return
}
