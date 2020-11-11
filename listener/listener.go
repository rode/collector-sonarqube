package listener

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/grafeas/grafeas/proto/v1beta1/common_go_proto"
	"github.com/grafeas/grafeas/proto/v1beta1/grafeas_go_proto"
	"github.com/grafeas/grafeas/proto/v1beta1/package_go_proto"
	"github.com/grafeas/grafeas/proto/v1beta1/vulnerability_go_proto"
	pb "github.com/liatrio/rode-api/proto/v1alpha1"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Event is...
type Event struct {
	TaskID      string            `json:"taskid"`
	Status      string            `json:"status"`
	AnalyzedAt  string            `json:"analyzedat"`
	GitCommit   string            `json:"revision"`
	Project     *Project          `json:"project"`
	QualityGate *QualityGate      `json:"qualityGate"`
	Branch      *Branch           `json:"branch"`
	Properties  map[string]string `json:"properties"`
}

// Branch is...
type Branch struct {
	Name   string `json:"name"`
	Type   string `json:"type"`
	IsMain bool   `json:"isMain"`
	URL    string `json:"url"`
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

const (
	address = "localhost:50051"
)

// ProcessEvent handles incoming webhook events
func ProcessEvent(w http.ResponseWriter, request *http.Request) {
	log.Print("Received SonarQube event")

	event := &Event{}
	if err := json.NewDecoder(request.Body).Decode(event); err != nil {
		w.WriteHeader(500)
		fmt.Fprintf(w, "Error reading webhook event: %s", err)
		return
	}

	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewRodeClient(conn)
	occurrences := []*grafeas_go_proto.Occurrence{}

	log.Printf("SonarQube Event Payload: [%+v]", event)
	log.Printf("SonarQube Event Project: [%+v]", event.Project)
	log.Printf("SonarQube Event Quality Gate: [%+v]", event.QualityGate)

	repo := getRepoFromSonar(event)

	for _, condition := range event.QualityGate.Conditions {
		log.Printf("SonarQube Event Quality Gate Condition: [%+v]", condition)
		occurrence := createQualityGateOccurrence(condition, repo)
		occurrences = append(occurrences, occurrence)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	response, err := c.BatchCreateOccurrences(ctx, &grafeas_go_proto.BatchCreateOccurrencesRequest{
		Occurrences: occurrences,
	})
	if err != nil {
		log.Fatalf("could not create occurrence: %v", err)
	}
	fmt.Printf("%#v\n", response)
	w.WriteHeader(200)
}

func getRepoFromSonar(event *Event) string {
	/*
		// Need to add logic to check if they are using developer or enterprise edition, but current API
		// only exposes this to admin users. Getting the resource URI is easier in the developer edition and is
		// not dependent on a value passed in from the project. It can be done like this:
		if isNotCommunityEdition(){
			repoString := fmt.Sprintf("%s:%s",event.Branch.URL,event.GitCommit)
			return repoString
		}
	*/

	repoString := fmt.Sprintf("%s:%s", event.Properties["sonar.analysis.resourceUriPrefix"], event.GitCommit)
	return repoString
}

func createQualityGateOccurrence(condition *Condition, repo string) *grafeas_go_proto.Occurrence {
	occurrence := &grafeas_go_proto.Occurrence{
		Name: condition.Metric,
		Resource: &grafeas_go_proto.Resource{
			Name: repo,
			Uri:  repo,
		},
		NoteName:    "projects/notes_project/notes/sonarqube",
		Kind:        common_go_proto.NoteKind_NOTE_KIND_UNSPECIFIED,
		Remediation: "test",
		CreateTime:  timestamppb.Now(),
		// To be changed when a proper occurrence type is determined
		Details: &grafeas_go_proto.Occurrence_Vulnerability{
			Vulnerability: &vulnerability_go_proto.Details{
				Type:             "test",
				Severity:         vulnerability_go_proto.Severity_CRITICAL,
				ShortDescription: "abc",
				LongDescription:  "abc123",
				RelatedUrls: []*common_go_proto.RelatedUrl{
					{
						Url:   "test",
						Label: "test",
					},
					{
						Url:   "test",
						Label: "test",
					},
				},
				EffectiveSeverity: vulnerability_go_proto.Severity_CRITICAL,
				PackageIssue: []*vulnerability_go_proto.PackageIssue{
					{
						SeverityName: "test",
						AffectedLocation: &vulnerability_go_proto.VulnerabilityLocation{
							CpeUri:  "test",
							Package: "test",
							Version: &package_go_proto.Version{
								Name:     "test",
								Revision: "test",
								Epoch:    35,
								Kind:     package_go_proto.Version_MINIMUM,
							},
						},
					},
				},
			},
		},
	}
	return occurrence
}
