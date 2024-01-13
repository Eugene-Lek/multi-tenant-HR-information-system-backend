package routes

import (
	"bufio"
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	_ "github.com/lib/pq"
	"github.com/pquerna/otp/totp"
)

func (s *IntegrationTestSuite) TestLoginPass() {
	type requestBody struct {
		TenantId string
		Email    string
		Password string
		Totp     string
	}

	totp, _ := totp.GenerateCode(s.defaultUser.TotpSecretKey, time.Now().UTC())
	reqBody := requestBody{
		TenantId: s.defaultUser.TenantId,
		Email:    s.defaultUser.Email,
		Password: "jU%q837d!QP7",
		Totp:     totp,
	}
	bodyBuf := new(bytes.Buffer)
	json.NewEncoder(bodyBuf).Encode(reqBody)

	r, err := http.NewRequest("POST", "/api/session", bodyBuf)
	if err != nil {
		log.Fatal(err)
	}

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, r)

	s.expectHttpStatus(w, 200)

	// Extract cookie from recorder and add it to the request
	cookieString := w.Header().Get("Set-Cookie")
	first, _, _ := strings.Cut(cookieString, ";")
	name, sessionId, _ := strings.Cut(first, "=")
	cookie := &http.Cookie{
		Name:  name,
		Value: sessionId,
	}
	r.AddCookie(cookie)

	session, err := s.sessionStore.Get(r, authSessionName)
	s.Equal(nil, err)
	if _, ok := session.Values["email"].(string); ok {
		s.Equal(s.defaultUser.Email, session.Values["email"].(string))
		s.Equal(s.defaultUser.TenantId, session.Values["tenantId"].(string))
	} else {
		s.Equal(true, ok, "Session should have been created")
	}

	reader := bufio.NewReader(s.logOutput)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"USER-AUTHORISED"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"SESSION-CREATED"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"USER-AUTHENTICATED"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"REQUEST-COMPLETED"`)
}

func (s *IntegrationTestSuite) TestLoginWrongCredentials() {
	type requestBody struct {
		TenantId string
		Email    string
		Password string
		Totp     string
	}

	totp, _ := totp.GenerateCode(s.defaultUser.TotpSecretKey, time.Now().UTC())

	tests := []struct {
		name  string
		input requestBody
	}{
		{
			"Login should fail because password is wrong",
			requestBody{
				TenantId: s.defaultUser.TenantId,
				Email:    s.defaultUser.Email,
				Password: "abcd1234!@#$%",
				Totp:     totp,
			},
		},
		{
			"Login should fail because totp is wrong",
			requestBody{
				TenantId: s.defaultUser.TenantId,
				Email:    s.defaultUser.Email,
				Password: "jU%q837d!QP7",
				Totp:     "123456",
			},
		},
		{
			"Login should fail because email is invalid",
			requestBody{
				TenantId: s.defaultUser.TenantId,
				Email:    "invalid@gmail.com",
				Password: "jU%q837d!QP7",
				Totp:     "123456",
			},
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			bodyBuf := new(bytes.Buffer)
			json.NewEncoder(bodyBuf).Encode(test.input)

			r, err := http.NewRequest("POST", "/api/session", bodyBuf)
			if err != nil {
				log.Fatal(err)
			}

			w := httptest.NewRecorder()
			s.router.ServeHTTP(w, r)

			s.expectHttpStatus(w, 401)
			s.expectErrorCode(w, "USER-UNAUTHENTICATED")

			// Extract cookie from recorder and add it to the request
			cookieString := w.Header().Get("Set-Cookie")
			first, _, _ := strings.Cut(cookieString, ";")
			name, sessionId, _ := strings.Cut(first, "=")
			cookie := &http.Cookie{
				Name:  name,
				Value: sessionId,
			}
			r.AddCookie(cookie)

			session, err := s.sessionStore.Get(r, authSessionName)
			s.Equal(nil, err)
			s.Equal("", session.ID, "Session should not have been created")

			reader := bufio.NewReader(s.logOutput)
			s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"USER-AUTHORISED"`)
			s.expectNextLogToContain(reader, `"level":"WARN"`, `"msg":"USER-UNAUTHENTICATED"`)
			s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"REQUEST-COMPLETED"`)
		})
	}
}

func (s *IntegrationTestSuite) TestLogout() {
	r, err := http.NewRequest("DELETE", "/api/session", nil)
	if err != nil {
		log.Fatal(err)
	}
	s.addSessionCookieToRequest(r, s.defaultUser.Id, s.defaultUser.TenantId, s.defaultUser.Email)

	// Duplicate the request so its cookie can be used for checking later
	reqCopy := *r

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, r)

	s.expectHttpStatus(w, 200)

	// Check that the session has been removed
	session, err := s.sessionStore.Get(&reqCopy, authSessionName)
	s.Equal(nil, err)

	_, ok := session.Values["email"].(string)
	s.Equal(false, ok, "Session should have been deleted")

	reader := bufio.NewReader(s.logOutput)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"USER-AUTHORISED"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"SESSION-DELETED"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"REQUEST-COMPLETED"`)
}

func (s *IntegrationTestSuite) TestLogoutTwice() {
	r, err := http.NewRequest("DELETE", "/api/session", nil)
	if err != nil {
		log.Fatal(err)
	}
	s.addSessionCookieToRequest(r, s.defaultUser.Id, s.defaultUser.TenantId, s.defaultUser.Email)

	// Copy the request to preserve its cookie for a 2nd logout attempt
	reqCopy := *r

	// Delete the session from the sessionStore via logout then clear the log
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, r)
	s.logOutput.Reset()

	w2 := httptest.NewRecorder()
	s.router.ServeHTTP(w2, &reqCopy)

	s.expectHttpStatus(w, 200)

	reader := bufio.NewReader(s.logOutput)
	s.expectNextLogToContain(reader, `"level":"WARN"`, `"msg":"DELETED-SESSION-USED"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"USER-AUTHORISED"`)
	s.expectNextLogToContain(reader, `"level":"WARN"`, `"msg":"SESSION-ALREADY-DELETED"`)
	s.expectNextLogToContain(reader, `"level":"INFO"`, `"msg":"REQUEST-COMPLETED"`)
}
