package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/gin-gonic/gin"
)

// CreateEventRequest represents the structure for creating a new event
type CreateEventRequest struct {
	Name        string            `json:"name" binding:"required" example:"my-scheduled-event"`
	Description string            `json:"description" example:"A scheduled event that runs daily"`
	Schedule    string            `json:"schedule" binding:"required" example:"0 12 * * ? *"` // cron expression
	Payload     map[string]string `json:"payload" example:"{\"key\":\"value\"}"`
}

// Event represents the response structure
type Event struct {
	Name        string            `json:"name" example:"my-scheduled-event"`
	Description string            `json:"description" example:"A scheduled event that runs daily"`
	Schedule    string            `json:"schedule" example:"0 12 * * ? *"`
	Payload     map[string]string `json:"payload" example:"{\"key\":\"value\"}"`
	CreatedAt   time.Time         `json:"created_at" example:"2024-03-20T12:00:00Z"`
}

// @Summary Create a new scheduled event
// @Description Create an AWS EventBridge rule with the provided schedule and payload
// @Tags events
// @Accept json
// @Produce json
// @Param event body controllers.CreateEventRequest true "Event data"
// @Success 201 {object} controllers.Event "Event created successfully"
// @Failure 400 {object} object "Invalid request data"
// @Failure 500 {object} object "Internal server error"
// @Router /events [post]
func CreateEvent(c *gin.Context) {
	// Get Auth0 user ID from the JWT claims
	claims, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Extract user ID from claims
	registeredClaims, ok := claims.(validator.RegisteredClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid claims format"})
		return
	}

	// Check if the user has the required Auth0 ID
	if registeredClaims.Subject != "auth0|68164b4c821b56fdc024b2dd" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req CreateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Load AWS configuration
	cfg, err := config.LoadDefaultConfig(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("unable to load SDK config: %v", err)})
		return
	}

	// Create EventBridge client
	client := eventbridge.NewFromConfig(cfg)

	// Convert payload to JSON string
	payloadJSON, err := json.Marshal(req.Payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to marshal payload: %v", err)})
		return
	}

	// Create rule input
	ruleInput := &eventbridge.PutRuleInput{
		Name:               aws.String(req.Name),
		Description:        aws.String(req.Description),
		ScheduleExpression: aws.String(fmt.Sprintf("cron(%s)", req.Schedule)),
		State:              types.RuleStateEnabled,
	}

	// Create the rule
	_, err = client.PutRule(c.Request.Context(), ruleInput)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to create rule: %v", err)})
		return
	}

	// Create target input
	targetInput := &eventbridge.PutTargetsInput{
		Rule: aws.String(req.Name),
		Targets: []types.Target{
			{
				Id:    aws.String(fmt.Sprintf("%s-target", req.Name)),
				Arn:   aws.String("arn:aws:lambda:us-east-1:069597727371:function:GoApiInfraStack-ProcessDataLambdaBB06D69F-fZPprYwiMNJa"), // Replace with your Lambda function ARN
				Input: aws.String(string(payloadJSON)),
			},
		},
	}

	// Create the target
	_, err = client.PutTargets(c.Request.Context(), targetInput)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to create target: %v", err)})
		return
	}

	// Create response
	event := Event{
		Name:        req.Name,
		Description: req.Description,
		Schedule:    req.Schedule,
		Payload:     req.Payload,
		CreatedAt:   time.Now(),
	}

	c.JSON(http.StatusCreated, event)
}
