package main

import (
	"encoding/json"
	"github"
	"net/http"
)

func handlePush(w http.ResponseWriter, r *http.Request) {
	var p github.HookMessage
	if err := json.Unmarshal([]byte(r.FormValue("payload")), &p); err != nil {
		errLogger.Println(err)
		return
	}

	logger.Println(p)
}
