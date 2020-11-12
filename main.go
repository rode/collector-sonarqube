package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	pb "github.com/liatrio/rode-api/proto/v1alpha1"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/liatrio/rode-collector-sonarqube/listener"
)

var (
	debug    bool
	port     int
	rodeHost string
)

func main() {
	flag.IntVar(&port, "port", 8080, "the port that the sonarqube collector should listen on")
	flag.BoolVar(&debug, "debug", false, "when set, debug mode will be enabled")
	flag.StringVar(&rodeHost, "rode-host", "localhost:50051", "the host to use to connect to rode")

	flag.Parse()

	logger, err := createLogger(debug)
	if err != nil {
		log.Fatalf("failed to create logger: %v", err)
	}

	conn, err := grpc.Dial(rodeHost, grpc.WithInsecure(), grpc.WithBlock())
	defer conn.Close()
	if err != nil {
		logger.Fatal("failed to establish grpc connection to Rode API", zap.NamedError("error", err))
	}

	rodeClient := pb.NewRodeClient(conn)

	l := listener.NewListener(logger.Named("listener"), rodeClient)

	http.HandleFunc("/webhook/event", l.ProcessEvent)
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) { fmt.Fprintf(w, "I'm healthy") })

	log.Println("Listening for SonarQube events")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func createLogger(debug bool) (*zap.Logger, error) {
	if debug {
		return zap.NewDevelopment()
	}

	return zap.NewProduction()
}
