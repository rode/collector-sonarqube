package listener

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// Event is...
type Event struct {
	TaskID      string       `json:"taskid"`
	Status      string       `json:"status"`
	AnalyzedAt  string       `json:"analyzedat"`
	GitCommit   string       `json:"revision"`
	Project     *Project     `json:"project"`
	QualityGate *QualityGate `json:"qualityGate"`
}

// Project is
type Project struct {
	Key  string `json:"key"`
	Name string `json:"name"`
	URL  string `json:"url"`
}

// QualityGate is...
type QualityGate struct {
	Conditions []*Condition `json:"conditions"`
	Name       string       `json:"name"`
	Status     string       `json:"status"`
}

// Condition is...
type Condition struct {
	ErrorThreshold string `json:"errorThreshold"`
	Metric         string `json:"metric"`
	OnLeakPeriod   bool   `json:"onLeakPeriod"`
	Operator       string `json:"operator"`
	Status         string `json:"status"`
}

// ProcessEvent handles incoming webhook events
func ProcessEvent(w http.ResponseWriter, request *http.Request) {
	log.Print("Received SonarQube event")

	event := &Event{}
	if err := json.NewDecoder(request.Body).Decode(event); err != nil {
		w.WriteHeader(500)
		fmt.Fprintf(w, "Error reading webhook event: %s", err)
		return
	}

	log.Printf("SonarQube Event Payload: [%+v]", event)
	log.Printf("SonarQube Event Project: [%+v]", event.Project)
	log.Printf("SonarQube Event Quality Gate: [%+v]", event.QualityGate)
	for _, condition := range event.QualityGate.Conditions {
		log.Printf("SonarQube Event Quality Gate Condition: [%+v]", condition)
	}

	w.WriteHeader(200)
}
