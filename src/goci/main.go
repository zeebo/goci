package main

import (
	"github.com/bmizerany/pat.go"
	"net/http"
	"os"
)

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

	//listen
	if err := http.ListenAndServe(":"+os.Getenv("PORT"), nil); err != nil {
		errLogger.Fatal(err)
	}
}
