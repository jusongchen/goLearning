package main

import (
	"log"
	"net/http"
)

func handler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("This is an example server.\n"))
}

func main() {
	http.HandleFunc("/", handler)
	log.Printf("About to listen on 8001.")
	err := http.ListenAndServeTLS(":8001", "cert.pem", "key.pem", nil)
	log.Fatal(err)
}
