package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/aws/smithy-go"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type mockAPIError struct {
	code    string
	message string
}

func (e mockAPIError) Error() string                 { return e.code + ": " + e.message }
func (e mockAPIError) ErrorCode() string             { return e.code }
func (e mockAPIError) ErrorMessage() string          { return e.message }
func (e mockAPIError) ErrorFault() smithy.ErrorFault { return smithy.FaultClient }

type mockEmailSender struct {
	SendEmailFunc func(ctx context.Context, req SendEmailRequest, auth0UserID string) (*SendEmailResponse, error)
}

func (m mockEmailSender) SendEmail(ctx context.Context, req SendEmailRequest, auth0UserID string) (*SendEmailResponse, error) {
	return m.SendEmailFunc(ctx, req, auth0UserID)
}

func setupEmailRouter(sender EmailSender) *gin.Engine {
	gin.SetMode(gin.TestMode)
	SetEmailSender(sender)
	router := gin.Default()
	router.Use(func(c *gin.Context) {
		c.Set("user", validator.RegisteredClaims{Subject: "auth0|sender"})
		c.Next()
	})
	router.POST("/email", SendEmail)
	return router
}

func TestSendEmail(t *testing.T) {
	router := setupEmailRouter(mockEmailSender{
		SendEmailFunc: func(ctx context.Context, req SendEmailRequest, auth0UserID string) (*SendEmailResponse, error) {
			assert.Equal(t, "auth0|sender", auth0UserID)
			assert.Equal(t, []string{"reader@example.com"}, req.To)
			assert.Equal(t, "Hello", req.Subject)
			assert.Equal(t, "A note from NoFeed.", req.Text)
			return &SendEmailResponse{
				From:      defaultEmailFromAddress,
				To:        req.To,
				MessageID: "message-123",
				Sent:      true,
			}, nil
		},
	})
	defer SetEmailSender(nil)

	body, _ := json.Marshal(SendEmailRequest{
		To:      []string{"Reader@Example.com"},
		Subject: " Hello ",
		Text:    " A note from NoFeed. ",
	})
	req := httptest.NewRequest(http.MethodPost, "/email", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusAccepted, w.Code)
	var res SendEmailResponse
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &res))
	assert.Equal(t, defaultEmailFromAddress, res.From)
	assert.Equal(t, "message-123", res.MessageID)
	assert.True(t, res.Sent)
}

func TestSendEmailRejectsInvalidRecipient(t *testing.T) {
	router := setupEmailRouter(mockEmailSender{})
	defer SetEmailSender(nil)

	body, _ := json.Marshal(SendEmailRequest{
		To:      []string{"not an email"},
		Subject: "Hello",
		Text:    "A note from NoFeed.",
	})
	req := httptest.NewRequest(http.MethodPost, "/email", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSendEmailRequiresBody(t *testing.T) {
	router := setupEmailRouter(mockEmailSender{})
	defer SetEmailSender(nil)

	body, _ := json.Marshal(SendEmailRequest{
		To:      []string{"reader@example.com"},
		Subject: "Hello",
	})
	req := httptest.NewRequest(http.MethodPost, "/email", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSendEmailReportsSenderFailure(t *testing.T) {
	router := setupEmailRouter(mockEmailSender{
		SendEmailFunc: func(ctx context.Context, req SendEmailRequest, auth0UserID string) (*SendEmailResponse, error) {
			return nil, errors.New("ses unavailable")
		},
	})
	defer SetEmailSender(nil)

	body, _ := json.Marshal(SendEmailRequest{
		To:      []string{"reader@example.com"},
		Subject: "Hello",
		Text:    "A note from NoFeed.",
	})
	req := httptest.NewRequest(http.MethodPost, "/email", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestSendEmailSurfacesSESMessageRejected(t *testing.T) {
	router := setupEmailRouter(mockEmailSender{
		SendEmailFunc: func(ctx context.Context, req SendEmailRequest, auth0UserID string) (*SendEmailResponse, error) {
			return nil, mockAPIError{
				code:    "MessageRejected",
				message: "Email address is not verified.",
			}
		},
	})
	defer SetEmailSender(nil)

	body, _ := json.Marshal(SendEmailRequest{
		To:      []string{"reader@example.com"},
		Subject: "Hello",
		Text:    "A note from NoFeed.",
	})
	req := httptest.NewRequest(http.MethodPost, "/email", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "SES rejected the message")
}
