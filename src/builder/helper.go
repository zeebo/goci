package builder

import "os"

func env(key, def string) string {
	if k := os.Getenv(key); k != "" {
		return k
	}
	return def
}
