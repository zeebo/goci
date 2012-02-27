package main

import (
	"encoding/json"
	"github"
	"net/http"
)

func handlePush(w http.ResponseWriter, r *http.Request) {
	dec := json.NewDecoder(r.Body)
	defer r.Body.Close()

	var p github.HookMessage
	if err := dec.Decode(&p); err != nil {
		errLogger.Printf("Error:", err)
		return
	}

	logger.Println(p)
}
