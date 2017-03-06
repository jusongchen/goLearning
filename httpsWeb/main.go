package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
)

func handler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("This is an example server.\n"))
}

func main() {
	port := 443
	TLS := true
	flag.IntVar(&port, "port", 443, "port to bind to")
	flag.BoolVar(&TLS, "TLS", true, "https with cert.pem and key.pem for SSL certificate")

	flag.Usage = func() {
		fmt.Println("Usage:")
		fmt.Printf("   %s [flags]  \n", os.Args[0])
		fmt.Println("Flags:")
		flag.PrintDefaults()
		fmt.Printf("\nThis programm opens cert.pem and key.pem for SSL certificate")
		os.Exit(-1)
	}

	flag.Parse()

	if flag.NArg() != 0 {
		flag.Usage()
	}
	portStr := fmt.Sprintf(":%d", port)

	http.HandleFunc("/", handler)

	log.Print("About to listen on ", portStr)
	if TLS {
		err := http.ListenAndServeTLS(portStr, "cert.pem", "key.pem", nil)
		log.Fatal(err)
	} else {
		err := http.ListenAndServe(portStr, nil)
		log.Fatal(err)
	}

}
