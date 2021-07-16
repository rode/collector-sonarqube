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
	"encoding/json"
	"errors"
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rode/collector-sonarqube/sonar"
	pb "github.com/rode/rode/proto/v1alpha1"
	"github.com/rode/rode/proto/v1alpha1fakes"
	"github.com/rode/rode/protodeps/grafeas/proto/v1beta1/common_go_proto"
	"github.com/rode/rode/protodeps/grafeas/proto/v1beta1/discovery_go_proto"
	"github.com/rode/rode/protodeps/grafeas/proto/v1beta1/grafeas_go_proto"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
)

var _ = Describe("listener", func() {
	var (
		rodeClient *v1alpha1fakes.FakeRodeClient
		listener   Listener
	)

	BeforeEach(func() {
		rodeClient = &v1alpha1fakes.FakeRodeClient{}
	})

	JustBeforeEach(func() {
		listener = NewListener(logger, rodeClient)
	})

	Context("ProcessEvent", func() {
		var (
			recorder           *httptest.ResponseRecorder
			expectedSonarEvent *sonar.Event
			expectedPayload    io.Reader

			expectedBatchCreateOccurrencesResponse *pb.BatchCreateOccurrencesResponse
			expectedBatchCreateOccurrencesError    error

			expectedTaskId            string
			expectedRevision          string
			expectedProjectUrl        string
			expectedQualityGateName   string
			expectedResourceUriPrefix string

			expectedNoteName        string
			expectedNote            *grafeas_go_proto.Note
			expectedCreateNoteError error
		)

		BeforeEach(func() {
			expectedPayload = nil
			recorder = httptest.NewRecorder()

			expectedTaskId = fake.LetterN(10)
			expectedRevision = fake.LetterN(10)
			expectedProjectUrl = fake.LetterN(10)
			expectedQualityGateName = fake.LetterN(10)
			expectedResourceUriPrefix = "git://" + fake.LetterN(10)

			expectedNoteName = fake.LetterN(10)
			expectedNote = &grafeas_go_proto.Note{
				Name: expectedNoteName,
			}
			expectedCreateNoteError = nil

			expectedBatchCreateOccurrencesResponse = &pb.BatchCreateOccurrencesResponse{}
			expectedBatchCreateOccurrencesError = nil
		})

		JustBeforeEach(func() {
			rodeClient.CreateNoteReturns(expectedNote, expectedCreateNoteError)
			rodeClient.BatchCreateOccurrencesReturns(expectedBatchCreateOccurrencesResponse, expectedBatchCreateOccurrencesError)

			var payload io.Reader
			if expectedPayload != nil {
				payload = expectedPayload
			} else {
				payload = structToJsonBody(expectedSonarEvent)
			}

			listener.ProcessEvent(recorder, httptest.NewRequest("POST", "/webhook/event", payload))
		})

		When("an invalid event is sent", func() {
			BeforeEach(func() {
				expectedPayload = strings.NewReader("invalid json")
			})

			It("should respond with an error", func() {
				Expect(recorder.Code).To(Equal(http.StatusInternalServerError))
			})

			It("should not make any request to rode", func() {
				Expect(rodeClient.CreateNoteCallCount()).To(Equal(0))
				Expect(rodeClient.BatchCreateOccurrencesCallCount()).To(Equal(0))
			})
		})

		When("an analysis occurs", func() {
			BeforeEach(func() {
				expectedSonarEvent = &sonar.Event{
					TaskId:     expectedTaskId,
					Status:     sonar.STATUS_SUCCESS,
					AnalysedAt: "2021-05-27T19:08:23+0000",
					Revision:   expectedRevision,
					Project: &sonar.Project{
						Key:  fake.LetterN(10),
						Name: fake.LetterN(10),
						URL:  expectedProjectUrl,
					},
					QualityGate: &sonar.QualityGate{
						Name:   expectedQualityGateName,
						Status: sonar.STATUS_OK,
					},
					Properties: map[string]string{
						resourceUriPrefixPropertyName: expectedResourceUriPrefix,
					},
				}
			})

			It("should create a note for the analysis", func() {
				Expect(rodeClient.CreateNoteCallCount()).To(Equal(1))

				_, createNoteRequest, _ := rodeClient.CreateNoteArgsForCall(0)

				Expect(createNoteRequest.NoteId).To(Equal(fmt.Sprintf("sonar-scan-%s", expectedTaskId)))

				Expect(createNoteRequest.Note.ShortDescription).To(Equal("SonarQube Analysis"))
				Expect(createNoteRequest.Note.LongDescription).To(Equal(fmt.Sprintf("SonarQube Analysis using %s Quality Gate", expectedQualityGateName)))
				Expect(createNoteRequest.Note.Kind).To(Equal(common_go_proto.NoteKind_DISCOVERY))

				Expect(createNoteRequest.Note.RelatedUrl).To(HaveLen(1))
				Expect(createNoteRequest.Note.RelatedUrl[0].Label).To(Equal("Project URL"))
				Expect(createNoteRequest.Note.RelatedUrl[0].Url).To(Equal(expectedProjectUrl))
			})

			It("should create discovery occurrences for the analysis", func() {
				Expect(rodeClient.BatchCreateOccurrencesCallCount()).To(Equal(1))

				_, batchCreateOccurrencesRequest, _ := rodeClient.BatchCreateOccurrencesArgsForCall(0)

				Expect(batchCreateOccurrencesRequest.Occurrences).To(HaveLen(2))

				scanStartOccurrence := batchCreateOccurrencesRequest.Occurrences[0]
				scanEndOccurrence := batchCreateOccurrencesRequest.Occurrences[1]

				Expect(scanStartOccurrence.Kind).To(Equal(common_go_proto.NoteKind_DISCOVERY))
				Expect(scanStartOccurrence.NoteName).To(Equal(expectedNoteName))
				Expect(scanStartOccurrence.Details.(*grafeas_go_proto.Occurrence_Discovered).Discovered.Discovered.AnalysisStatus).To(Equal(discovery_go_proto.Discovered_SCANNING))
				Expect(scanStartOccurrence.Resource.Uri).To(Equal(fmt.Sprintf("%s@%s", expectedResourceUriPrefix, expectedRevision)))

				Expect(scanEndOccurrence.Kind).To(Equal(common_go_proto.NoteKind_DISCOVERY))
				Expect(scanEndOccurrence.NoteName).To(Equal(expectedNoteName))
				Expect(scanEndOccurrence.Details.(*grafeas_go_proto.Occurrence_Discovered).Discovered.Discovered.AnalysisStatus).To(Equal(discovery_go_proto.Discovered_FINISHED_SUCCESS))
				Expect(scanEndOccurrence.Resource.Uri).To(Equal(fmt.Sprintf("%s@%s", expectedResourceUriPrefix, expectedRevision)))
			})

			It("should respond with a 200", func() {
				Expect(recorder.Code).To(Equal(http.StatusOK))
			})

			When("the analysis fails", func() {
				BeforeEach(func() {
					expectedSonarEvent.Status = sonar.STATUS_FAILED
					expectedSonarEvent.QualityGate = nil
				})

				It("should indicate the failure in the discovery occurrence", func() {
					Expect(rodeClient.BatchCreateOccurrencesCallCount()).To(Equal(1))

					_, batchCreateOccurrencesRequest, _ := rodeClient.BatchCreateOccurrencesArgsForCall(0)

					Expect(batchCreateOccurrencesRequest.Occurrences).To(HaveLen(2))

					scanStartOccurrence := batchCreateOccurrencesRequest.Occurrences[0]
					scanEndOccurrence := batchCreateOccurrencesRequest.Occurrences[1]

					Expect(scanStartOccurrence.Kind).To(Equal(common_go_proto.NoteKind_DISCOVERY))
					Expect(scanStartOccurrence.NoteName).To(Equal(expectedNoteName))
					Expect(scanStartOccurrence.Details.(*grafeas_go_proto.Occurrence_Discovered).Discovered.Discovered.AnalysisStatus).To(Equal(discovery_go_proto.Discovered_SCANNING))
					Expect(scanStartOccurrence.Resource.Uri).To(Equal(fmt.Sprintf("%s@%s", expectedResourceUriPrefix, expectedRevision)))

					Expect(scanEndOccurrence.Kind).To(Equal(common_go_proto.NoteKind_DISCOVERY))
					Expect(scanEndOccurrence.NoteName).To(Equal(expectedNoteName))
					Expect(scanEndOccurrence.Details.(*grafeas_go_proto.Occurrence_Discovered).Discovered.Discovered.AnalysisStatus).To(Equal(discovery_go_proto.Discovered_FINISHED_FAILED))
					Expect(scanEndOccurrence.Resource.Uri).To(Equal(fmt.Sprintf("%s@%s", expectedResourceUriPrefix, expectedRevision)))
				})
			})

			When("an unexpected payload is received from sonar", func() {
				BeforeEach(func() {
					expectedSonarEvent.QualityGate = nil
				})

				It("should respond with a 500", func() {
					Expect(recorder.Code).To(Equal(http.StatusInternalServerError))
				})

				It("should not create a note", func() {
					Expect(rodeClient.CreateNoteCallCount()).To(Equal(0))
				})

				It("should not create occurrences", func() {
					Expect(rodeClient.BatchCreateOccurrencesCallCount()).To(Equal(0))
				})
			})

			When("the analysis timestamp is invalid", func() {
				BeforeEach(func() {
					expectedSonarEvent.AnalysedAt = fake.LetterN(10)
				})

				It("should respond with a 500", func() {
					Expect(recorder.Code).To(Equal(http.StatusInternalServerError))
				})

				It("should not create occurrences", func() {
					Expect(rodeClient.BatchCreateOccurrencesCallCount()).To(Equal(0))
				})
			})

			When("the resource uri prefix property is missing", func() {
				BeforeEach(func() {
					delete(expectedSonarEvent.Properties, resourceUriPrefixPropertyName)
				})

				It("should respond with a 200 (user error)", func() {
					Expect(recorder.Code).To(Equal(http.StatusOK))
				})

				It("should not create a note", func() {
					Expect(rodeClient.CreateNoteCallCount()).To(Equal(0))
				})

				It("should not create occurrences", func() {
					Expect(rodeClient.BatchCreateOccurrencesCallCount()).To(Equal(0))
				})
			})

			When("the git:// prefix is not specified", func() {
				BeforeEach(func() {
					expectedSonarEvent.Properties[resourceUriPrefixPropertyName] = strings.TrimPrefix(expectedResourceUriPrefix, "git://")
				})

				It("should be added automatically", func() {
					_, batchCreateOccurrencesRequest, _ := rodeClient.BatchCreateOccurrencesArgsForCall(0)

					scanStartOccurrence := batchCreateOccurrencesRequest.Occurrences[0]
					scanEndOccurrence := batchCreateOccurrencesRequest.Occurrences[1]

					Expect(scanStartOccurrence.Resource.Uri).To(Equal(fmt.Sprintf("%s@%s", expectedResourceUriPrefix, expectedRevision)))
					Expect(scanEndOccurrence.Resource.Uri).To(Equal(fmt.Sprintf("%s@%s", expectedResourceUriPrefix, expectedRevision)))
				})
			})

			When("creating the note fails", func() {
				BeforeEach(func() {
					expectedCreateNoteError = errors.New("error creating note")
				})

				It("should respond with a 500", func() {
					Expect(recorder.Code).To(Equal(http.StatusInternalServerError))
				})

				It("should not create occurrences", func() {
					Expect(rodeClient.BatchCreateOccurrencesCallCount()).To(Equal(0))
				})
			})
		})
	})
})

func structToJsonBody(i interface{}) io.ReadCloser {
	b, err := json.Marshal(i)
	Expect(err).ToNot(HaveOccurred())

	return ioutil.NopCloser(strings.NewReader(string(b)))
}
