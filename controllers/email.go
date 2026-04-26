package controllers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/mail"
	"strings"
	"unicode/utf8"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
	"github.com/dat1010/go-api/utils"
	"github.com/gin-gonic/gin"
)

const emailFromAddress = "testing@nofeed.zone"

var (
	emailSender EmailSender = sesEmailSender{}

	errEmailSubjectRequired = errors.New("subject is required")
	errEmailBodyRequired    = errors.New("text or html is required")
	errEmailToRequired      = errors.New("at least one recipient is required")
)

// EmailSender sends email for authenticated users.
type EmailSender interface {
	SendEmail(ctx context.Context, req SendEmailRequest, auth0UserID string) (*SendEmailResponse, error)
}

type sesEmailSender struct{}

// SendEmailRequest represents an email request from a logged-in user.
type SendEmailRequest struct {
	To      []string `json:"to" binding:"required" example:"friend@example.com"`
	Subject string   `json:"subject" binding:"required" example:"NoFeed test"`
	Text    string   `json:"text" example:"Plain text body"`
	HTML    string   `json:"html" example:"<p>HTML body</p>"`
}

// SendEmailResponse represents a successful email send.
type SendEmailResponse struct {
	From      string   `json:"from"`
	To        []string `json:"to"`
	MessageID string   `json:"message_id"`
	Sent      bool     `json:"sent"`
}

// SetEmailSender swaps the email sender implementation, mainly for tests.
func SetEmailSender(sender EmailSender) {
	if sender == nil {
		emailSender = sesEmailSender{}
		return
	}
	emailSender = sender
}

func (s sesEmailSender) SendEmail(ctx context.Context, req SendEmailRequest, auth0UserID string) (*SendEmailResponse, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config: %w", err)
	}

	client := sesv2.NewFromConfig(cfg)
	body := &types.Body{}
	if req.Text != "" {
		body.Text = &types.Content{
			Charset: aws.String("UTF-8"),
			Data:    aws.String(req.Text),
		}
	}
	if req.HTML != "" {
		body.Html = &types.Content{
			Charset: aws.String("UTF-8"),
			Data:    aws.String(req.HTML),
		}
	}

	out, err := client.SendEmail(ctx, &sesv2.SendEmailInput{
		FromEmailAddress: aws.String(emailFromAddress),
		Destination: &types.Destination{
			ToAddresses: req.To,
		},
		Content: &types.EmailContent{
			Simple: &types.Message{
				Subject: &types.Content{
					Charset: aws.String("UTF-8"),
					Data:    aws.String(req.Subject),
				},
				Body: body,
			},
		},
		EmailTags: []types.MessageTag{
			{
				Name:  aws.String("auth0_user_id"),
				Value: aws.String(sanitizeSESTagValue(auth0UserID)),
			},
		},
	})
	if err != nil {
		return nil, err
	}

	messageID := ""
	if out.MessageId != nil {
		messageID = *out.MessageId
	}
	return &SendEmailResponse{
		From:      emailFromAddress,
		To:        req.To,
		MessageID: messageID,
		Sent:      true,
	}, nil
}

// SendEmail godoc
// @Summary      Send email
// @Description  Send an email from testing@nofeed.zone for the authenticated user.
// @Tags         notifications
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        request body controllers.SendEmailRequest true "Email payload"
// @Success      202 {object} controllers.SendEmailResponse
// @Failure      400 {object} object "Invalid request"
// @Failure      401 {object} object "Unauthorized"
// @Failure      500 {object} object "Internal server error"
// @Router       /email [post]
func SendEmail(c *gin.Context) {
	auth0UserID, ok := utils.GetAuth0UserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req SendEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := validateEmailRequest(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := emailSender.SendEmail(c.Request.Context(), req, auth0UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to send email: %v", err)})
		return
	}
	c.JSON(http.StatusAccepted, res)
}

func validateEmailRequest(req *SendEmailRequest) error {
	req.Subject = strings.TrimSpace(req.Subject)
	req.Text = strings.TrimSpace(req.Text)
	req.HTML = strings.TrimSpace(req.HTML)
	req.To = normalizeRecipients(req.To)

	if len(req.To) == 0 {
		return errEmailToRequired
	}
	if req.Subject == "" {
		return errEmailSubjectRequired
	}
	if utf8.RuneCountInString(req.Subject) > 200 {
		return errors.New("subject must be 200 characters or fewer")
	}
	if req.Text == "" && req.HTML == "" {
		return errEmailBodyRequired
	}
	if len(req.To) > 10 {
		return errors.New("too many recipients; maximum is 10")
	}
	for _, recipient := range req.To {
		parsed, err := mail.ParseAddress(recipient)
		if err != nil || parsed.Address != recipient {
			return fmt.Errorf("invalid recipient: %s", recipient)
		}
	}
	return nil
}

func normalizeRecipients(recipients []string) []string {
	seen := make(map[string]bool)
	normalized := make([]string, 0, len(recipients))
	for _, recipient := range recipients {
		trimmed := strings.ToLower(strings.TrimSpace(recipient))
		if trimmed == "" || seen[trimmed] {
			continue
		}
		seen[trimmed] = true
		normalized = append(normalized, trimmed)
	}
	return normalized
}

func sanitizeSESTagValue(value string) string {
	replacer := strings.NewReplacer("|", "-", ":", "-", "/", "-", "\\", "-")
	sanitized := replacer.Replace(value)
	if sanitized == "" {
		return "unknown"
	}
	if len(sanitized) > 256 {
		return sanitized[:256]
	}
	return sanitized
}
