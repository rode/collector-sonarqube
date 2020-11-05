package main

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
)

// Webhook is...
type Webhook struct {
	Ctx          context.Context
	Sonar        SonarQube
	Key          string
	Name         string
	Organization string
	Project      string
	URL          string
	Secret       string
}

// WebhookResponse is...
type WebhookResponse struct {
	Key    string `json:"key"`
	Name   string `json:"name"`
	URL    string `json:"url"`
	Secret string `json:"secret,omitempty"`
}

// WebhookResponseItem is...
type WebhookResponseItem struct {
	Webhook WebhookResponse `json:"webhook"`
}

// WebhookResponseList is...
type WebhookResponseList struct {
	Webhooks []WebhookResponse `json:"webhooks"`
}

// WebhookEvent is...
type WebhookEvent struct {
	TaskID      string       `json:"taskid"`
	Status      string       `json:"status"`
	AnalyzedAt  string       `json:"analyzedat"`
	GitCommit   string       `json:"revision"`
	Project     *Project     `json:"project"`
	QualityGate *QualityGate `json:"qualityGate"`
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
	OnLeakPeriod   string `json:"onLeakPeriod"`
	Operator       string `json:"operator"`
	Status         string `json:"status"`
}

// NewWebhook constructor for Webhook struct
func NewWebhook(ctx context.Context, sonar SonarQube, name string, url string, organization string, project string) (Webhook, error) {
	if name == "" {
		return Webhook{}, errors.New("Webhook name is required")
	}
	if url == "" {
		return Webhook{}, errors.New("Webhook URL is required")
	}
	hash := md5.Sum([]byte(url))
	name = name + "-" + hex.EncodeToString(hash[:])[:8]
	return Webhook{
		Ctx:          ctx,
		Sonar:        sonar,
		Name:         name,
		Organization: organization,
		Project:      project,
		URL:          url}, nil
}

// Create creates a new webhook
func (w *Webhook) Create() error {
	_ = w.clean()

	client := &http.Client{}
	params := url.Values{}
	params.Add("name", w.Name)
	params.Add("url", w.URL)
	if w.Organization != "" {
		params.Add("organization", w.Organization)
	}
	if w.Project != "" {
		params.Add("project", w.Project)
	}
	w.Secret = w.createSecret()
	params.Add("secret", w.Secret)
	request, err := w.Sonar.Request("POST", "api/webhooks/create", strings.NewReader(params.Encode()))
	if err != nil {
		return fmt.Errorf("Error creating webhook request: %s", err)
	}
	response, err := client.Do(request)
	err = checkResponse(response, err, 200)
	if err != nil {
		return err
	}

	var responseBody = &WebhookResponseItem{}
	if err := json.NewDecoder(response.Body).Decode(responseBody); err != nil {
		return fmt.Errorf("Error decoding create webhook response: %s", err)
	}
	w.Key = responseBody.Webhook.Key
	w.Secret = responseBody.Webhook.Secret

	return nil
}

// Delete deletes rode-collector webhook using existing object key
func (w *Webhook) Delete() error {
	err := w.deleteByKey(w.Key)
	if err != nil {
		return fmt.Errorf("Error deleting webhook: %s", err)
	}

	return nil
}

// ProcessEvent handles incoming webhook events
func (w *Webhook) ProcessEvent(writer http.ResponseWriter, request *http.Request) {
	log.Print("Received SonarQube event")

	event := &WebhookEvent{}
	json.NewDecoder(request.Body).Decode(event)
	log.Printf("SonarQube Event Payload: [%+v]", event)

	writer.WriteHeader(200)
}

// deleteBykey deletes an webhook using the specified key identifer
func (w *Webhook) deleteByKey(key string) error {
	client := &http.Client{}
	params := url.Values{}
	params.Add("webhook", key)
	request, err := w.Sonar.Request("POST", "api/webhooks/delete", strings.NewReader(params.Encode()))
	if err != nil {
		return fmt.Errorf("Error createing delete webhook request: %s", err)
	}

	resp, err := client.Do(request)
	err = checkResponse(resp, err, 201)
	if err != nil {
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
	client := &http.Client{}
	params := url.Values{}

	params.Add("name", w.Name)
	request, err := w.Sonar.Request("GET", "api/webhooks/list", strings.NewReader(params.Encode()))
	if err != nil {
		log.Printf("Error creating list webhook request: %s\n", err)
		return err
	}

	response, err := client.Do(request)
	err = checkResponse(response, err, 200)
	if err != nil {
		return err
	}

	body := &WebhookResponseList{}
	if err := json.NewDecoder(response.Body).Decode(body); err != nil {

		return fmt.Errorf("Error decoding webhook list: %+v", response)
	}

	for _, webhook := range body.Webhooks {
		if webhook.Name == w.Name {
			err = w.deleteByKey(webhook.Key)
			if err != nil {
				log.Printf("Error deleting stale webhook: %s", err)
			}
		}
	}

	return nil
}

func checkResponse(resp *http.Response, err error, okStatusCode int) error {
	if err != nil {
		return fmt.Errorf("Failed to make request: [%s]", err)
	}

	if resp.StatusCode != okStatusCode {
		return fmt.Errorf("Unexpected response: [%d]", resp.StatusCode)
	}

	return nil
}
