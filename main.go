package main

import (
	"context"
	"flag"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"
)

var (
	flagCollectorURL  string
	flagSonarURL      string = "http://localhost:9000"
	flagSonarUsername string = "admin"
	flagSonarPassword string = "admin"
	flagSonarToken    string
)

func stringEnv(name string, defaultValue string) string {
	if value, ok := os.LookupEnv(name); ok {
		return value
	}
	return defaultValue
}

func init() {
	rand.Seed(time.Now().UnixNano())

	flag.StringVar(&flagCollectorURL, "url", os.Getenv("URL"), "Collector webhook URL")
	flag.StringVar(&flagSonarURL, "sonar-url", stringEnv("SONAR_URL", flagSonarURL), "URL of SonarQube host")
	flag.StringVar(&flagSonarUsername, "sonar-username", stringEnv("SONAR_USERNAME", flagSonarUsername), "Username for SonarQube host")
	flag.StringVar(&flagSonarPassword, "sonar-password", stringEnv("SONAR_PASSWORD", flagSonarPassword), "Password for SonarQube host")
	flag.StringVar(&flagSonarToken, "sonar-token", os.Getenv("SONAR_TOKEN"), "Token for SonarQube host")
}

func main() {
	flag.Parse()
	ctx := context.Background()
	var auth Auth
	if flagSonarToken != "" {
		auth = &AuthToken{token: flagSonarToken}
	} else if flagSonarUsername != "" && flagSonarPassword != "" {
		auth = &AuthBasic{username: flagSonarUsername, password: flagSonarPassword}
	} else {
		log.Fatal("Must include SonarQube username and password or token")
	}
	client := SonarQubeClient{ctx: ctx, url: flagSonarURL, auth: auth}
	wh := NewWebhook(ctx, client, "rode-collector", flagCollectorURL, "", "")
	wh.Create()
	defer wh.Delete()

	http.HandleFunc("/webhook/event", wh.ProcessEvent)

	log.Println("Listening for SonarQube events")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
