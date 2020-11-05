package main

import (
	"context"
	"encoding/base64"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Sonarqube", func() {
	Describe("Set username/password Auth method for SonarQube connection", func() {
		Context("With a username and password", func() {
			It("Should set BasicAuth header", func() {
				var (
					auth     Auth
					username string
					password string
					request  *http.Request
				)
				username = "fakename"
				password = "fakepass"
				auth = &AuthBasic{Username: username, Password: password}

				expectedAuthValue := base64.StdEncoding.EncodeToString([]byte(username + ":" + password))

				request = httptest.NewRequest("GET", "http://fake.com", strings.NewReader("Fake body"))
				auth.Inject(request)

				Expect(request.Header.Get("Authorization")).To(Equal("Basic " + expectedAuthValue))
			})
		})

		Context("With a token", func() {
			It("Should set BasicAuth header", func() {
				var (
					auth    Auth
					token   string
					request *http.Request
				)
				token = "faketoken"
				auth = &AuthToken{Token: token}

				expectedAuthValue := base64.StdEncoding.EncodeToString([]byte(token + ":"))

				request = httptest.NewRequest("GET", "http://fake.com", strings.NewReader("Fake body"))
				auth.Inject(request)

				Expect(request.Header.Get("Authorization")).To(Equal("Basic " + expectedAuthValue))
			})
		})
	})

	Describe("Create a new SonarQube request object", func() {
		var (
			method    string
			path      string
			body      io.Reader
			ctx       context.Context
			sonarQube *SonarQubeClient
		)

		BeforeEach(func() {
			method = "GET"
			path = "http://fake.com"
			body = strings.NewReader("Simple response")
			sonarQube = &SonarQubeClient{
				Ctx:  ctx,
				Url:  path,
				Auth: &AuthToken{Token: "xxx"},
			}
		})

		Context("With an invalid URL", func() {
			It("Should return an error", func() {
				path = "http//invalid.url.com"

				req, err := sonarQube.Request(method, path, body)

				Expect(err).To(HaveOccurred(), "Invalid request URL should return an error")
				Expect(req).To(Equal(&http.Request{}))
			})
		})
	})
})
