package main

import (
	"fmt"
	"log"
	"net/http"
)

func webserver(tokenstream chan string) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		token_code := q.Get("code")
		log.Println(token_code)
		tokenstream <- token_code
		fmt.Fprintln(w, "You may close this tab")
	})
	log.Panic(http.ListenAndServe(":80", nil))
}
