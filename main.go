package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"liatr.io/rode-collector-sonarqube/listener"
)

func main() {
	flag.Parse()
	http.HandleFunc("/webhook/event", listener.ProcessEvent)
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) { fmt.Fprintf(w, "I'm healthy") })

	log.Println("Listening for SonarQube events")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
