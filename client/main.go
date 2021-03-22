package main

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"os"

	"golang.org/x/net/http2"
)

func main() {

	client := http.Client{
		Transport: &http2.Transport{},
	}

	resp, err := client.Post("https://localhost:3000", "application/json", bytes.NewReader([]byte("{\"message\":\"hello\"}")))
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	file, err := os.Create("image.png")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	io.Copy(file, resp.Body)
	log.Printf("Protocol Version: %s\n", resp.Proto)
}
