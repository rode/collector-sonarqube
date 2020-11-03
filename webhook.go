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

// deleteBykey deletes an webhook using the specified key identifer
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
