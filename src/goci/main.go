package main

import (
	"github.com/bmizerany/pat.go"
	"net/http"
	"os"
	"path/filepath"
)

func herokuTmpDir(w http.ResponseWriter, req *http.Request) {
	dir := os.TempDir()
	logger.Println(dir)
	logger.Println("Mkdir", os.Mkdir(filepath.Join(dir, "foo"), 0777))
}

func main() {
	m := pat.New()
	m.Get("/debug/:id", http.HandlerFunc(debugDatabase))

	http.Handle("/debug/", m)
	http.HandleFunc("/push", handlePush)
	http.HandleFunc("/tmpdir", herokuTmpDir)
	if err := http.ListenAndServe(":"+os.Getenv("PORT"), nil); err != nil {
		errLogger.Fatal(err)
	}
}
