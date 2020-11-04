package main

import (
	"context"
	"net/http"

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

		It("Should return an error if a name is not set", func() {
			webhookName = ""
			_, err := NewWebhook(ctx, client, webhookName, webhookURL, sonarqubeOrg, sonarqubeProject)
			Expect(err).To(HaveOccurred())
		})

		It("Should return an error if a URL is not set", func() {
			webhookURL = ""
			_, err := NewWebhook(ctx, client, webhookName, webhookURL, sonarqubeOrg, sonarqubeProject)
			Expect(err).To(HaveOccurred())
		})

		It("Should return a Webhook object when called with both required parameters", func() {
			wh, err := NewWebhook(ctx, client, webhookName, webhookURL, sonarqubeOrg, sonarqubeProject)
			Expect(err).ToNot(HaveOccurred())
			Expect(wh.Sonar).To(Equal(client))
			Expect(wh.Name).To(ContainSubstring(webhookName))
			Expect(wh.Organization).To(Equal(sonarqubeOrg))
			Expect(wh.Project).To(Equal(sonarqubeProject))
			Expect(wh.URL).To(Equal(webhookURL))
		})
	})

	Describe("Create new SonarQube webhook", func() {
		var (
			webhookResponse *WebhookResponse
			webhookList     []WebhookResponse
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

		})

		Context("And the SonarQube webhook is not created", func() {
			It("Should return a webhook request creation error", func() {
				httpmock.RegisterResponder("POST", webhookURL+"/api/webhooks/create",
					func(req *http.Request) (*http.Response, error) {
						return httpmock.NewJsonResponse(500, "")
					},
				)

				wh, _ := NewWebhook(ctx, client, webhookName, webhookURL, sonarqubeOrg, sonarqubeProject)
				err := wh.Create()
				Expect(err).To(HaveOccurred(), "Bad response from webhook creation should return an error")
			})
		})

		Context("And the SonarQube webhook is created", func() {
			It("Should return a webhook request creation error", func() {
				httpmock.RegisterResponder("POST", webhookURL+"/api/webhooks/create",
					func(req *http.Request) (*http.Response, error) {
						return httpmock.NewJsonResponse(200, webhookResponse)
					},
				)

				wh, _ := NewWebhook(ctx, client, webhookName, webhookURL, sonarqubeOrg, sonarqubeProject)
				err := wh.Create()
				Expect(err).ToNot(HaveOccurred(), "Bad response from webhook creation should return an error")
			})
		})

		Context("And the response body does not match WebhookResponseItem", func() {
			It("Should return a decoding error", func() {
				httpmock.RegisterResponder("POST", webhookURL+"/api/webhooks/create",
					func(req *http.Request) (*http.Response, error) {
						return httpmock.NewJsonResponse(200, "")
					},
				)

				wh, _ := NewWebhook(ctx, client, webhookName, webhookURL, sonarqubeOrg, sonarqubeProject)
				err := wh.Create()
				Expect(err).To(HaveOccurred(), "Bad response body from webhook creation should return an decoding error")
			})
		})
	})
})
