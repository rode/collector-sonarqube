package main

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strings"

	"github.com/grafeas/grafeas/proto/v1beta1/common_go_proto"
	"github.com/grafeas/grafeas/proto/v1beta1/grafeas_go_proto"
	"github.com/grafeas/grafeas/proto/v1beta1/package_go_proto"
	"github.com/grafeas/grafeas/proto/v1beta1/vulnerability_go_proto"
	"google.golang.org/protobuf/types/known/timestamppb"
	"liatr.io/rode-collector-sonarqube/client"
)

type Webhook struct {
	ctx          context.Context
	sonar        SonarQubeClient
	key          string
	name         string
	organization string
	project      string
	url          string
	secret       string
}

type WebhookResponse struct {
	Key    string `json:"key"`
	Name   string `json:"name"`
	URL    string `json:"url"`
	Secret string `json:"secret,omitempty"`
}

type WebhookResponseItem struct {
	Webhook WebhookResponse `json:"webhook"`
}

type WebhookResponseList struct {
	Webhooks []WebhookResponse `json:"webhooks"`
}

type WebhookEvent struct {
	TaskId      string       `json:"taskid"`
	Status      string       `json:"status"`
	AnalyzedAt  string       `json:"analyzedat"`
	GitCommit   string       `json:"revision"`
	Project     *Project     `json:"project"`
	QualityGate *QualityGate `json:"qualityGate"`
}

type Project struct {
	Key  string `json:"key"`
	Name string `json:"name"`
	Url  string `json:"url"`
}

type QualityGate struct {
	Conditions []*Condition `json:"conditions"`
	Name       string       `json:"name"`
	Status     string       `json:"status"`
}

type Condition struct {
	ErrorThreshold string `json:"errorThreshold`
	Metric         string `json:"metric"`
	OnLeakPeriod   string `json:"onLeakPeriod"`
	Operator       string `json:"operator"`
	Status         string `json:"status"`
}

// NewWebhook constructor for Webhook struct
func NewWebhook(ctx context.Context, sonar SonarQubeClient, name string, url string, organization string, project string) Webhook {
	if name == "" {
		log.Fatal("Webhook name is required")
	}
	if url == "" {
		log.Fatal("Webhook URL is required")
	}
	hash := md5.Sum([]byte(url))
	name = name + "-" + hex.EncodeToString(hash[:])[:8]
	return Webhook{
		ctx:          ctx,
		sonar:        sonar,
		name:         name,
		organization: organization,
		project:      project,
		url:          url}
}

// Create creates a new webhook
func (w *Webhook) Create() {
	_ = w.clean()

	client := http.Client{}
	params := url.Values{}
	params.Add("name", w.name)
	params.Add("url", w.url)
	if w.organization != "" {
		params.Add("organization", w.organization)
	}
	if w.project != "" {
		params.Add("project", w.project)
	}
	w.secret = w.createSecret()
	params.Add("secret", w.secret)
	request, err := w.sonar.Request("POST", "api/webhooks/create", strings.NewReader(params.Encode()))
	if err != nil {
		log.Fatalf("Error creating webhook request: %s", err)
	}
	response, err := client.Do(request)
	if err != nil {
		log.Fatalf("Error creating SonarQube webhook: %s", err)
	}
	var responseBody = &WebhookResponseItem{}
	dec := json.NewDecoder(response.Body)
	err = dec.Decode(responseBody)
	if err != nil {
		log.Fatalf("Error decoding create webhook response: %s", err)
	}
	w.key = responseBody.Webhook.Key
	w.secret = responseBody.Webhook.Secret
}

// Delete deletes rode-collector webhook using existing object key
func (w *Webhook) Delete() {
	err := w.deleteByKey(w.key)
	if err != nil {
		log.Fatalf("Error deleting webhook: %s", err)
	}
}

// ProcessEvent handles incoming webhook events
func (w *Webhook) ProcessEvent(writer http.ResponseWriter, request *http.Request) {
	log.Print("Received SonarQube event")

	event := &WebhookEvent{}
	json.NewDecoder(request.Body).Decode(event)
	log.Printf("SonarQube Event Payload: [%+v]", event)
	log.Printf("SonarQube Event Project: [%+v]", event.Project)
	log.Printf("SonarQube Event Quality Gate: [%+v]", event.QualityGate)
	log.Printf("SonarQube Event Quality Gate: [%+v]", event.QualityGate.Conditions[0])

	client := client.RodeClient{URL: "localhost:50051"}
	client.SendOccurrences([]*grafeas_go_proto.Occurrence{w.createQualityGateOccurrence(event)})

	writer.WriteHeader(200)
}

// deleteByKey deletes an webhook using the specified key identifer
func (w *Webhook) deleteByKey(key string) error {
	client := http.Client{}
	params := url.Values{}
	params.Add("webhook", key)
	request, err := w.sonar.Request("POST", "api/webhooks/delete", strings.NewReader(params.Encode()))
	if err != nil {
		log.Printf("Error createing delete webhook request: %s", err)
		return err
	}
	_, err = client.Do(request)
	if err != nil {
		log.Printf("Error deleting webhook: %s", err)
		return err
	}
	return nil
}

// createSecret generates a random string to use as web secret
func (w *Webhook) createSecret() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	b := make([]byte, 200)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// clean deletes stale webhooks from SonarQube
func (w *Webhook) clean() error {
	client := http.Client{}
	params := url.Values{}
	params.Add("name", w.name)
	request, err := w.sonar.Request("GET", "api/webhooks/list", strings.NewReader(params.Encode()))
	if err != nil {
		log.Printf("Error creating list webhook request: %s\n", err)
		return err
	}
	response, err := client.Do(request)
	if err != nil {
		log.Printf("Error fetching webhooks from SonarQube: %s", err)
	}
	body := &WebhookResponseList{}

	if err := json.NewDecoder(response.Body).Decode(body); err != nil {
		log.Printf("Error decoding webhook list: %s\n", err)
		return err
	}
	for _, webhook := range body.Webhooks {
		if webhook.Name == w.name {
			err = w.deleteByKey(webhook.Key)
			if err != nil {
				log.Printf("Error deleting stale webhook: %s", err)
			}
		}
	}
	return err
}

func (w *Webhook) createQualityGateOccurrence(webhook *WebhookEvent) *grafeas_go_proto.Occurrence {
	occurrence := &grafeas_go_proto.Occurrence{
		Name: "abc",
		Resource: &grafeas_go_proto.Resource{
			Name: "testResource",
			Uri:  "test",
		},
		NoteName:    "projects/abc/notes/123",
		Kind:        common_go_proto.NoteKind_VULNERABILITY,
		Remediation: "test",
		CreateTime:  timestamppb.Now(),
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
