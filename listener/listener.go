// Copyright 2021 The Rode Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package listener

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/rode/rode/protodeps/grafeas/proto/v1beta1/discovery_go_proto"
	"net/http"
	"strings"
	"time"

	"github.com/rode/collector-sonarqube/sonar"
	"go.uber.org/zap"

	pb "github.com/rode/rode/proto/v1alpha1"
	"github.com/rode/rode/protodeps/grafeas/proto/v1beta1/common_go_proto"
	"github.com/rode/rode/protodeps/grafeas/proto/v1beta1/grafeas_go_proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	resourceUriPrefixPropertyName = "sonar.analysis.resourceUriPrefix"
)

type listener struct {
	rodeClient pb.RodeClient
	logger     *zap.Logger
}

type Listener interface {
	ProcessEvent(http.ResponseWriter, *http.Request)
}

func NewListener(logger *zap.Logger, client pb.RodeClient) Listener {
	return &listener{
		rodeClient: client,
		logger:     logger,
	}
}

// ProcessEvent handles incoming webhook events
func (l *listener) ProcessEvent(w http.ResponseWriter, request *http.Request) {
	log := l.logger.Named("ProcessEvent")

	event := &sonar.Event{}
	if err := json.NewDecoder(request.Body).Decode(event); err != nil {
		log.Error("error reading webhook event", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log = log.With(zap.Any("event", event))
	log.Debug("received sonarqube event")

	resourceUri, err := getResourceUriFromEvent(event)
	if err != nil {
		log.Error("error getting resource uri from event", zap.Error(err))
		// there's no point in responding with an error to sonar, as this is a user error
		w.WriteHeader(http.StatusOK)
		return
	}

	// allow for one minute to process this event
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// create a note to represent the sonar analysis
	noteName, err := l.createNoteForEvent(ctx, event)
	if err != nil {
		log.Error("error creating note for analysis", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// create occurrences for sonar analysis
	response, err := l.createOccurrencesForEvent(ctx, event, resourceUri, noteName)
	if err != nil {
		log.Error("error creating occurrences for event", zap.Error(err))
		w.WriteHeader(500)
		return
	}

	log.Debug("response payload", zap.Any("response", response.GetOccurrences()))
	w.WriteHeader(200)
}

// getResourceUriFromEvent parses the received event and returns a resource URI that can be referenced in occurrences.
// Eventually, this needs to check if the community edition or the developer edition of sonar is being used. The events
// sent from the developer edition should contain information about the repository URL that can be used to construct the
// resource URI. Unfortunately, the community edition requires that an extra property "resourceUriPrefix" is sent with the
// scan. We'll throw an error here if this property isn't present.
func getResourceUriFromEvent(event *sonar.Event) (string, error) {
	prefix, ok := event.Properties[resourceUriPrefixPropertyName]
	if !ok {
		return "", fmt.Errorf("expected event to contain the resource uri prefix. please run the scanner with the \"-D%s\" option", resourceUriPrefixPropertyName)
	}

	// add the expected "git://" prefix if it isn't provided
	if !strings.HasPrefix(prefix, "git://") {
		prefix = fmt.Sprintf("git://%s", prefix)
	}

	return fmt.Sprintf("%s@%s", prefix, event.Revision), nil
}

// createNoteForEvent creates a note that represents the sonar analysis.
func (l *listener) createNoteForEvent(ctx context.Context, event *sonar.Event) (string, error) {
	var longDescription string
	if event.Status == sonar.STATUS_FAILED {
		longDescription = "Failed SonarQube Analysis"
	} else if event.Status == sonar.STATUS_SUCCESS && event.QualityGate != nil {
		longDescription = fmt.Sprintf("SonarQube Analysis using %s Quality Gate", event.QualityGate.Name)
	} else {
		return "", errors.New("unexpected event payload, unable to compute note for event")
	}

	note, err := l.rodeClient.CreateNote(ctx, &pb.CreateNoteRequest{
		Note: &grafeas_go_proto.Note{
			ShortDescription: "SonarQube Analysis",
			LongDescription:  longDescription,
			Kind:             common_go_proto.NoteKind_DISCOVERY,
			RelatedUrl: []*common_go_proto.RelatedUrl{
				{
					Label: "Project URL",
					Url:   event.Project.URL,
				},
			},
			Type: &grafeas_go_proto.Note_Discovery{
				Discovery: &discovery_go_proto.Discovery{
					// in the future, this should reference the new static analysis note kind
					AnalysisKind: common_go_proto.NoteKind_VULNERABILITY,
				},
			},
		},
		NoteId: fmt.Sprintf("sonar-scan-%s", event.TaskId),
	})
	if err != nil {
		return "", err
	}

	return note.Name, nil
}

// createOccurrencesForEvent creates occurrences based on the received sonar event. We use discovery occurrences here
// due to the lack of a better occurrence type. We also misuse the discovery analysis status, such that "FAILED" is
// equivalent to a failing quality gate, rather than the analysis as a whole failing. This will be revisited with the
// addition of a new static analysis occurrence type.
func (l *listener) createOccurrencesForEvent(ctx context.Context, event *sonar.Event, resourceUri, noteName string) (*pb.BatchCreateOccurrencesResponse, error) {
	timestamp, err := eventTimestamp(event)
	if err != nil {
		return nil, err
	}

	status := discovery_go_proto.Discovered_FINISHED_FAILED
	if event.Status == sonar.STATUS_SUCCESS && event.QualityGate != nil && event.QualityGate.Status == sonar.STATUS_OK {
		status = discovery_go_proto.Discovered_FINISHED_SUCCESS
	}

	return l.rodeClient.BatchCreateOccurrences(ctx, &pb.BatchCreateOccurrencesRequest{
		Occurrences: []*grafeas_go_proto.Occurrence{
			{
				Resource: &grafeas_go_proto.Resource{
					Uri: resourceUri,
				},
				NoteName:   noteName,
				Kind:       common_go_proto.NoteKind_DISCOVERY,
				CreateTime: timestamp,
				Details: &grafeas_go_proto.Occurrence_Discovered{
					Discovered: &discovery_go_proto.Details{
						Discovered: &discovery_go_proto.Discovered{
							ContinuousAnalysis: discovery_go_proto.Discovered_CONTINUOUS_ANALYSIS_UNSPECIFIED,
							AnalysisStatus:     discovery_go_proto.Discovered_SCANNING,
						},
					},
				},
			},
			{
				Resource: &grafeas_go_proto.Resource{
					Uri: resourceUri,
				},
				NoteName:   noteName,
				Kind:       common_go_proto.NoteKind_DISCOVERY,
				CreateTime: timestamp,
				Details: &grafeas_go_proto.Occurrence_Discovered{
					Discovered: &discovery_go_proto.Details{
						Discovered: &discovery_go_proto.Discovered{
							ContinuousAnalysis: discovery_go_proto.Discovered_CONTINUOUS_ANALYSIS_UNSPECIFIED,
							AnalysisStatus:     status,
						},
					},
				},
			},
		},
	})
}

func eventTimestamp(event *sonar.Event) (*timestamppb.Timestamp, error) {
	timestamp, err := time.Parse("2006-01-02T15:04:05+0000", event.AnalysedAt)
	if err != nil {
		return nil, err
	}

	return timestamppb.New(timestamp), nil
}
