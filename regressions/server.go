package main

import (
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
)

func JsonResponseHandler(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "{\"data\":[\"string\"]}")
}

func EchoHandler(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		http.Error(w, "unable to read body", 400)
		return
	}

	w.Write(data)
}

func EchoHeadersHandler(w http.ResponseWriter, req *http.Request) {
	keys := make([]string, 0)
	for key, _ := range req.Header {
		keys = append(keys, key)
	}

	sort.Strings(keys)
	for _, key := range keys {
		value := req.Header[key][0]
		data := []byte(key + ": " + value + "\n")
		w.Write(data)
	}
}

func main() {
	http.HandleFunc("/404", http.NotFound)
	http.HandleFunc("/json-response", JsonResponseHandler)
	http.HandleFunc("/echo-body", EchoHandler)
	http.HandleFunc("/echo-headers", EchoHeadersHandler)
	log.Fatal(http.ListenAndServe(":12345", nil))
}
