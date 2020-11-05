package main

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/jarcoal/httpmock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Webhook", func() {
	var (
		webhookName      string
		webhookURL       string
		sonarqubeOrg     string
		sonarqubeProject string
		ctx              context.Context
		client           *SonarQubeClient
	)

	BeforeEach(func() {
		webhookName = "test-rode-collector"
		webhookURL = "http://rode-collector.fake.com/webhook/event"
		sonarqubeOrg = "fake-org"
		sonarqubeProject = "fake-project"
		ctx = context.Background()
		client = &SonarQubeClient{Ctx: ctx, Url: webhookURL, Auth: &AuthToken{Token: "test-token"}}
	})

	Describe("Creating a new Webhook object", func() {
		Context("Without providing a webhook name", func() {
			It("Should return an error", func() {
				webhookName = ""

				wh, err := NewWebhook(
					ctx,
					client,
					webhookName,
					webhookURL,
					sonarqubeOrg,
					sonarqubeProject,
				)
				Expect(err).To(HaveOccurred(), "Webhooks must have a name")
				Expect(wh).To(Equal(Webhook{}), "Object should be empty")
			})
		})

		It("Without providing a webhook URL", func() {
			webhookURL = ""

			wh, err := NewWebhook(
				ctx,
				client,
				webhookName,
				webhookURL,
				sonarqubeOrg,
				sonarqubeProject,
			)
			Expect(err).To(HaveOccurred(), "Webhooks must have a URL")
			Expect(wh).To(Equal(Webhook{}), "Object should be empty")
		})

		It("With all required parameters", func() {
			hash := md5.Sum([]byte(webhookURL))
			hashedName := webhookName + "-" + hex.EncodeToString(hash[:])[:8]

			wh, err := NewWebhook(
				ctx,
				client,
				webhookName,
				webhookURL,
				sonarqubeOrg,
				sonarqubeProject,
			)
			Expect(err).ToNot(HaveOccurred())
			Expect(wh.Sonar).To(Equal(client))
			Expect(wh.Name).To(Equal(hashedName), "Webhook name should be hashed with URL")
			Expect(wh.Organization).To(Equal(sonarqubeOrg))
			Expect(wh.Project).To(Equal(sonarqubeProject))
			Expect(wh.URL).To(Equal(webhookURL))
		})
	})

	Describe("Create new SonarQube webhook", func() {
		var (
			webhookResponse *WebhookResponse
			webhookList     []WebhookResponse
			wh              Webhook
			err             error
		)

		BeforeEach(func() {
			webhookResponse = &WebhookResponse{Key: "xxx", Name: webhookName, URL: webhookURL}
			webhookList = []WebhookResponse{}
			webhookList = append(webhookList, *webhookResponse)

			httpmock.RegisterResponder("GET", webhookURL+"/api/webhooks/list",
				func(req *http.Request) (*http.Response, error) {
					return httpmock.NewJsonResponse(200, &WebhookResponseList{Webhooks: webhookList})
				},
			)

			wh, err = NewWebhook(
				ctx,
				client,
				webhookName,
				webhookURL,
				sonarqubeOrg,
				sonarqubeProject,
			)
		})

		It("With a bad response from the SonarQube host", func() {
			httpmock.RegisterResponder("POST", webhookURL+"/api/webhooks/create",
				func(req *http.Request) (*http.Response, error) {
					return httpmock.NewJsonResponse(500, "")
				},
			)

			err = wh.Create()

			Expect(err).To(HaveOccurred(), "Bad response from webhook creation should return an error")
		})

		It("Without a valid connection to SonarQube", func() {
			httpmock.RegisterResponder("POST", webhookURL+"/api/webhooks/create",
				func(req *http.Request) (*http.Response, error) {
					return nil, errors.New("Cannot reach SonarQube host")
				},
			)

			err = wh.Create()
			Expect(err).To(HaveOccurred(), "Bad connection should return an error")
		})

		Context("And the SonarQube webhook creation succeeds", func() {
			It("Should return a webhook request creation error", func() {
				httpmock.RegisterResponder("POST", webhookURL+"/api/webhooks/create",
					func(req *http.Request) (*http.Response, error) {
						return httpmock.NewJsonResponse(200, webhookResponse)
					},
				)
				err = wh.Create()
				Expect(err).ToNot(HaveOccurred(), "Bad response from webhook creation should return an error")
			})
		})

		It("With an invalid response body", func() {
			httpmock.RegisterResponder("POST", webhookURL+"/api/webhooks/create",
				func(req *http.Request) (*http.Response, error) {
					return httpmock.NewJsonResponse(200, "")
				},
			)
			err = wh.Create()
			Expect(err).To(HaveOccurred(), "Bad response body from webhook creation should return an decoding error")
		})
	})

	Describe("Deleting SonarQube Webhook", func() {
		var (
			webhookResponse *WebhookResponse
			webhookList     []WebhookResponse
			wh              Webhook
			err             error
		)

		BeforeEach(func() {
			webhookResponse = &WebhookResponse{Key: "xxx", Name: webhookName, URL: webhookURL}
			webhookList = []WebhookResponse{}
			webhookList = append(webhookList, *webhookResponse)

			httpmock.RegisterResponder("GET", webhookURL+"/api/webhooks/list",
				func(req *http.Request) (*http.Response, error) {
					return httpmock.NewJsonResponse(200, &WebhookResponseList{Webhooks: webhookList})
				},
			)

			wh, _ = NewWebhook(
				ctx,
				client,
				webhookName,
				webhookURL,
				sonarqubeOrg,
				sonarqubeProject,
			)
		})

		Context("Without a valid connection to SonarQube", func() {
			It("Should return with an error", func() {
				httpmock.RegisterResponder("POST", webhookURL+"/api/webhooks/delete",
					func(req *http.Request) (*http.Response, error) {
						return nil, errors.New("Cannot reach SonarQube host")
					},
				)

				err = wh.deleteByKey("invalid_key")
				Expect(err).To(HaveOccurred(), "Bad connection should return an error")
			})
		})

		Context("With an invalid webhook key", func() {
			It("Should return an error", func() {
				httpmock.RegisterResponder("POST", webhookURL+"/api/webhooks/delete",
					func(req *http.Request) (*http.Response, error) {
						return httpmock.NewJsonResponse(404, "")
					},
				)

				err = wh.deleteByKey("invalid_key")
				Expect(err).To(HaveOccurred(), "Using a non-existent key should return cause an error")
			})
		})

		Context("With an valid webhook key", func() {
			It("Should return nil", func() {
				httpmock.RegisterResponder("POST", webhookURL+"/api/webhooks/delete",
					func(req *http.Request) (*http.Response, error) {
						return httpmock.NewJsonResponse(201, "")
					},
				)

				err = wh.deleteByKey("valid_key")
				Expect(err).ToNot(HaveOccurred(), "Using a valid key should return nil")
			})
		})

		Context("Delete Method called from control loop with an invalid key", func() {
			It("Without a valid webhook key", func() {
				wh.Key = "invalid_key"
				httpmock.RegisterResponder("POST", webhookURL+"/api/webhooks/delete",
					func(req *http.Request) (*http.Response, error) {
						return httpmock.NewJsonResponse(404, "")
					},
				)

				err = wh.Delete()
				Expect(err).To(HaveOccurred(), "Using a non-existent key should return cause an error")
			})
		})

		Context("Delete Method called from control loop with an valid key", func() {
			It("With a valid webhook key", func() {
				wh.Key = "valid_key"
				httpmock.RegisterResponder("POST", webhookURL+"/api/webhooks/delete",
					func(req *http.Request) (*http.Response, error) {
						return httpmock.NewJsonResponse(201, "")
					},
				)

				err = wh.Delete()
				Expect(err).ToNot(HaveOccurred(), "Using a valid key should return nil")
			})
		})
	})

	Describe("Clean pre-existing Rode Webhooks from SonarQube", func() {
		var (
			webhookResponse *WebhookResponse
			webhookList     []WebhookResponse
			wh              Webhook
			err             error
		)

		BeforeEach(func() {
			webhookResponse = &WebhookResponse{Key: "xxx", Name: webhookName, URL: webhookURL}
			webhookList = []WebhookResponse{}
			webhookList = append(webhookList, *webhookResponse)

			wh, _ = NewWebhook(
				ctx,
				client,
				webhookName,
				webhookURL,
				sonarqubeOrg,
				sonarqubeProject,
			)
		})

		Context("With a bad connection to SonarQube", func() {
			It("Should return an error", func() {
				httpmock.RegisterResponder("GET", webhookURL+"/api/webhooks/list",
					func(req *http.Request) (*http.Response, error) {
						return nil, errors.New("Cannot reach SonarQube host")
					},
				)

				err = wh.clean()
				Expect(err).To(HaveOccurred(), "Bad connection should return an error")
			})
		})

		Context("With a bad response from SonarQube", func() {
			It("Should return an error", func() {
				httpmock.RegisterResponder("GET", webhookURL+"/api/webhooks/list",
					func(req *http.Request) (*http.Response, error) {
						return httpmock.NewJsonResponse(500, "")
					},
				)

				err = wh.clean()
				Expect(err).To(HaveOccurred(), "Bad response should return an error")
			})
		})

		Context("With an empty response from SonarQube", func() {
			It("Should return an error", func() {
				httpmock.RegisterResponder("GET", webhookURL+"/api/webhooks/list",
					func(req *http.Request) (*http.Response, error) {
						return httpmock.NewJsonResponse(200, "")
					},
				)

				err = wh.clean()
				Expect(err).To(HaveOccurred(), "Bad response should return an error")
			})
		})

		Context("With an empty response from SonarQube", func() {
			It("Should return an error", func() {
				httpmock.RegisterResponder("GET", webhookURL+"/api/webhooks/list",
					func(req *http.Request) (*http.Response, error) {
						return httpmock.NewJsonResponse(200, "")
					},
				)

				err = wh.clean()
				Expect(err).To(HaveOccurred(), "Bad response should return an error")
			})
		})

		Context("With a unsuccessful webhook cleanup ", func() {
			It("Should not return an error", func() {
				webhookResponse := &WebhookResponse{Key: "xxx", Name: wh.Name, URL: webhookURL}
				webhookList := []WebhookResponse{}
				webhookList = append(webhookList, *webhookResponse)

				httpmock.RegisterResponder("GET", webhookURL+"/api/webhooks/list",
					func(req *http.Request) (*http.Response, error) {
						return httpmock.NewJsonResponse(200, &WebhookResponseList{Webhooks: webhookList})
					},
				)
				httpmock.RegisterResponder("POST", webhookURL+"/api/webhooks/delete",
					func(req *http.Request) (*http.Response, error) {
						return httpmock.NewJsonResponse(401, "")
					},
				)

				err = wh.clean()
				Expect(err).ToNot(HaveOccurred(), "We don't need to handle failed webhook cleanups")
			})
		})
	})

	Describe("Process Webhook event", func() {
		var wh Webhook

		BeforeEach(func() {
			wh, _ = NewWebhook(
				ctx,
				client,
				webhookName,
				webhookURL,
				sonarqubeOrg,
				sonarqubeProject,
			)
		})

		Context("With new valid event", func() {
			It("Should not error out", func() {
				req, _ := http.NewRequest("POST", "/webhook/event", strings.NewReader("Event!"))
				rr := httptest.NewRecorder()
				handler := http.HandlerFunc(wh.ProcessEvent)

				handler.ServeHTTP(rr, req)
			})
		})
	})
})
