package listener

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Listener", func() {

	var (
		listener Listener
	)

	//var generalEvent *Event

	BeforeEach(func() {
		//generalEvent = &Event{
		//	TaskID:     "AXW39ga8rPJdWZ8bmqZS",
		//	Status:     "SUCCESS",
		//	AnalyzedAt: "2020-11-11T15:39:05+0000",
		//	GitCommit:  "e4834abbbd161241694224b3b91950b3d504a3a3",
		//	Project: &Project{
		//		Key:  "org.springframework.nanotrader:springtrader-marketSummary",
		//		Name: "springtrader-marketSummary",
		//		URL:  "http://localhost:9000/dashboard?id=org.springframework.nanotrader%3Aspringtrader-marketSummary",
		//	},
		//	Properties: map[string]string{
		//		"sonar.analysis.resourceUriPrefix": "https://github.com/liatrio/springtrader-marketsummary-java",
		//	},
		//}

		listener = NewListener(logger)
	})
	Context("Determining Resource URI", func() {
		When("using Sonarqube Community Edition", func() {
			It("should be based on a passed in resource uri prefix", func() {
				Expect(getRepoFromSonar(generalEvent)).To(Equal("https://github.com/liatrio/springtrader-marketsummary-java:e4834abbbd161241694224b3b91950b3d504a3a3"))
			})
		})

	})

	Describe("Process incoming event", func() {
		Context("With new valid event", func() {
			It("Should not error out", func() {
				var (
					coverageCondition     *Condition
					qualityGateConditions []*Condition
				)

				coverageCondition = &Condition{
					ErrorThreshold: "80",
					Metric:         "new_coverage",
					OnLeakPeriod:   true,
					Operator:       "LESS_THAN",
					Status:         "OK",
				}
				qualityGateConditions = append(qualityGateConditions, coverageCondition)
				event := &Event{
					TaskID:     "xxx",
					Status:     "OK",
					AnalyzedAt: "2016-11-18T10:46:28+0100",
					GitCommit:  "c739069ec7105e01303e8b3065a81141aad9f129",
					Project: &Project{
						Key:  "testproject",
						Name: "Test Project",
						URL:  "https://mycompany.com/sonarqube/dashboard?id=myproject",
					},
					QualityGate: &QualityGate{
						Conditions: qualityGateConditions,
						Name:       "SonarQube way",
						Status:     "OK",
					},
				}
				body, _ := json.Marshal(event)
				req, _ := http.NewRequest("POST", "/webhook/event", bytes.NewBuffer(body))
				rr := httptest.NewRecorder()
				handler := http.HandlerFunc(ProcessEvent)

				handler.ServeHTTP(rr, req)
				Expect(rr.Result().StatusCode).To(Equal(200))
			})
		})

		Context("With invalid event", func() {
			It("Should return a bad response", func() {
				req, _ := http.NewRequest("POST", "/webhook/event", strings.NewReader("Bad object"))
				rr := httptest.NewRecorder()
				handler := http.HandlerFunc(ProcessEvent)

				handler.ServeHTTP(rr, req)
				Expect(rr.Code).To(Equal(500))
				Expect(rr.Body.String()).To(ContainSubstring("Error reading webhook event"))
			})
		})
	})
})
