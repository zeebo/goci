package main

import (
	"encoding/json"
	"github.com/bmizerany/pat.go"
	"io"
	"net/http"
	"os"
)

func writeFully(w io.Writer, data []byte) (err error) {
	var m int
	for n := 0; n < len(data); n += m {
		m, err = w.Write(data[n:])
		if err != nil {
			return
		}
	}
	return
}

func handleRecentResults(w http.ResponseWriter, req *http.Request) {
	data, _ := json.MarshalIndent(<-recentResults, "", "\t")
	writeFully(w, data)
}

func handleQueue(w http.ResponseWriter, req *http.Request) {
	data, _ := json.MarshalIndent(<-currentQueue, "", "\t")
	writeFully(w, data)
}

func main() {
	//set up our database connection
	if err := setupDatabase(); err != nil {
		errLogger.Fatal(err)
	}
	go resultInsert()

	//set up our routing system
	m := pat.New()
	m.Get("/debug/:id", http.HandlerFunc(debugDatabase))
	http.Handle("/debug/", m)
	http.HandleFunc("/github/hook", handleGithubPush)

	http.HandleFunc("/recent", handleRecentResults)
	http.HandleFunc("/queue", handleQueue)

	//listen
	if err := http.ListenAndServe(":"+os.Getenv("PORT"), nil); err != nil {
		errLogger.Fatal(err)
	}
}
