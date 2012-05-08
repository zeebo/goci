package builder

import (
	"crypto/sha1"
	"fmt"
)

func hash(in string) (out string) {
	h := sha1.New()
	h.Write([]byte(in))
	out = fmt.Sprintf("%x", h.Sum(nil)[:5])
	return
}
