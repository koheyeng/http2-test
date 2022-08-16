package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Hello world!!")
	})

	err := http.ListenAndServeTLS(":3000", "localhost.pem", "localhost-key.pem", nil)
	if err != nil {
		log.Fatal(err)
	}
}
