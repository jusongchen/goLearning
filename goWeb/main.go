package main

import (
	"fmt"
	"net/http"

	"github.com/jusongchen/goLearning/goWeb/views"
)

var (
	index, contact, msg *views.View
)

func main() {
	index = views.NewView("bootstrap", "views/index.gohtml")
	contact = views.NewView("bootstrap", "views/contacts.gohtml")
	msg = views.NewView("bootstrap", "views/message.gohtml")

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/contact", contactHandler)
	http.HandleFunc("/message", msgHandler)
	http.ListenAndServe(":3000", nil)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	index.Render(w, nil)
}

func contactHandler(w http.ResponseWriter, r *http.Request) {
	contact.Render(w, nil)
}

func msgHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("msgHandler")
	msg.Render(w, nil)
}
