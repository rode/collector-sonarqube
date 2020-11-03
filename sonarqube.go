package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
)

type SonarQubeClient struct {
	ctx  context.Context
	url  string
	auth Auth
}

type Auth interface {
	Inject(*http.Request)
}

type AuthBasic struct {
	username string
	password string
}

type AuthToken struct {
	token string
}

func (a *AuthBasic) Inject(request *http.Request) {
	request.SetBasicAuth(a.username, a.username)
}

func (a *AuthToken) Inject(request *http.Request) {
	request.SetBasicAuth(a.token, "")
}

// Request creates an http.Request for SonarQube API
func (s *SonarQubeClient) Request(method string, path string, body io.Reader) (*http.Request, error) {
	url := fmt.Sprintf("%s/%s", s.url, path)
	request, err := http.NewRequestWithContext(s.ctx, method, url, body)
	if err != nil {
		log.Printf("Error creating SonarQube request (METHOD:%s, URL:%s, BODY:%s ERROR:%s)", method, url, body, err)
		return nil, err
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	s.auth.Inject(request)
	return request, nil
}
