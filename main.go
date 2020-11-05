package main

import (
	"log"
	"net/http"

	"liatr.io/rode-collector-sonarqube/listener"
)

func main() {
	http.HandleFunc("/webhook/event", listener.ProcessEvent)

	log.Println("Listening for SonarQube events")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
