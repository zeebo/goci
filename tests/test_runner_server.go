package main

import (
	"io"
	"log"
	"net/http"
	"os"
)

func dump(w http.ResponseWriter, req *http.Request) {
	io.Copy(os.Stdout, req.Body)
}

func main() {
	http.HandleFunc("/resp", dump)
	http.Handle("/", http.FileServer(http.Dir(".")))

	if err := http.ListenAndServe(":"+os.Getenv("PORT"), nil); err != nil {
		log.Fatal(err)
	}
}
