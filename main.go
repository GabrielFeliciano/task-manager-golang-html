package main

import (
	"html/template"
	"log"
	"net/http"
	"time"
)

func main() {
	helloMsg, _ := template.New("Hello message").Parse(`
		<h1>Hello boy! Calling this at {{.}}<h1>
	`)

	// http
	http.HandleFunc("/api/msg", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			now := time.Now().Format(time.RFC3339)
			helloMsg.Execute(w, now)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	log.Fatal(http.ListenAndServe(":3050", nil))
}
