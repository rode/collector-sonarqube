package listener

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/liatrio/rode-collector-sonarqube/sonar"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Listener", func() {

	var (
		listener     Listener
		rodeClient   *mockRodeClient
		generalEvent *sonar.Event
	)

	BeforeEach(func() {
		var (
			coverageCondition     *sonar.Condition
			qualityGateConditions []*sonar.Condition
		)

		coverageCondition = &sonar.Condition{
			ErrorThreshold: "80",
			Metric:         "new_coverage",
			OnLeakPeriod:   true,
			Operator:       "LESS_THAN",
			Status:         "OK",
		}
		qualityGateConditions = append(qualityGateConditions, coverageCondition)

		generalEvent = &sonar.Event{
			TaskID:     "AXW39ga8rPJdWZ8bmqZS",
			Status:     "SUCCESS",
			AnalyzedAt: "2020-11-11T15:39:05+0000",
			GitCommit:  "e4834abbbd161241694224b3b91950b3d504a3a3",
			Project: &sonar.Project{
				Key:  "org.springframework.nanotrader:springtrader-marketSummary",
				Name: "springtrader-marketSummary",
				URL:  "http://localhost:9000/dashboard?id=org.springframework.nanotrader%3Aspringtrader-marketSummary",
			},
			Properties: map[string]string{
				"sonar.analysis.resourceUriPrefix": "https://github.com/liatrio/springtrader-marketsummary-java",
			},
			QualityGate: &sonar.QualityGate{
				Conditions: qualityGateConditions,
				Name:       "SonarQube way",
				Status:     "OK",
			}}

		rodeClient = &mockRodeClient{}
		listener = NewListener(logger, rodeClient)
		rodeClient.expectedError = nil
	})
	Context("Determining Resource URI", func() {
		When("using Sonarqube Community Edition", func() {
			It("should be based on a passed in resource uri prefix", func() {
				Expect(getRepoFromSonar(generalEvent)).To(Equal("https://github.com/liatrio/springtrader-marketsummary-java:e4834abbbd161241694224b3b91950b3d504a3a3"))
			})
		})

	})
	Context("Processing incoming event", func() {
		When("using a valid event", func() {
			It("should not error out", func() {
				body, _ := json.Marshal(generalEvent)
				req, _ := http.NewRequest("POST", "/webhook/event", bytes.NewBuffer(body))
				rr := httptest.NewRecorder()
				handler := http.HandlerFunc(listener.ProcessEvent)

				handler.ServeHTTP(rr, req)
				Expect(rr.Result().StatusCode).To(Equal(200))
			})
		})

		When("using an invalid event", func() {
			It("Should return a bad response", func() {
				req, _ := http.NewRequest("POST", "/webhook/event", strings.NewReader("Bad object"))
				rr := httptest.NewRecorder()
				handler := http.HandlerFunc(listener.ProcessEvent)

				handler.ServeHTTP(rr, req)
				Expect(rr.Code).To(Equal(500))
				Expect(rr.Body.String()).To(ContainSubstring("error reading webhook event"))
			})
		})

		When("failing to create occurrences", func() {
			It("Should return a bad response", func() {
				rodeClient.expectedError = (errors.New("FAILED"))

				body, _ := json.Marshal(generalEvent)
				req, _ := http.NewRequest("POST", "/webhook/event", bytes.NewBuffer(body))
				rr := httptest.NewRecorder()
				handler := http.HandlerFunc(listener.ProcessEvent)

				handler.ServeHTTP(rr, req)
				Expect(rr.Code).To(Equal(500))
			})
		})
	})
})
