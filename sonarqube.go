package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

type SonarQubeClient struct {
	Ctx  context.Context
	Url  string
	Auth Auth
}

type SonarQube interface {
	Request(method string, path string, body io.Reader) (*http.Request, error)
}

type Auth interface {
	Inject(*http.Request)
}

type AuthBasic struct {
	Username string
	Password string
}

type AuthToken struct {
	Token string
}

func (a *AuthBasic) Inject(request *http.Request) {
	request.SetBasicAuth(a.Username, a.Password)
}

func (a *AuthToken) Inject(request *http.Request) {
	request.SetBasicAuth(a.Token, "")
}

// Request creates an http.Request for SonarQube API
func (s *SonarQubeClient) Request(method string, path string, body io.Reader) (*http.Request, error) {
	url := fmt.Sprintf("%s/%s", s.Url, path)
	request, err := http.NewRequestWithContext(s.Ctx, method, url, body)
	if err != nil {
		return &http.Request{}, fmt.Errorf("Error creating SonarQube request (METHOD:%s, URL:%s, BODY:%s ERROR:%s)", method, url, body, err)
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	s.Auth.Inject(request)
	return request, nil
}
