package main

import (
	"io"
	"log"
	"net/http"
)

func JsonResponseHandler(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "{\"data\":[\"string\"]}")
}

func main() {
	http.HandleFunc("/404", http.NotFound)
	http.HandleFunc("/json-response", JsonResponseHandler)
	log.Fatal(http.ListenAndServe(":12345", nil))
}
