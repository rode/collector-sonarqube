package main

import (
	"flag"
	"log"
	"net/http"

	"liatr.io/rode-collector-sonarqube/listener"
)

func main() {
	flag.Parse()
	http.HandleFunc("/webhook/event", listener.ProcessEvent)

	log.Println("Listening for SonarQube events")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
