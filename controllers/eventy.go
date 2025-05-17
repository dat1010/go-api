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
	"github.com/gin-gonic/gin"
)

// CreateEventRequest represents the structure for creating a new event
type CreateEventRequest struct {
	Name        string            `json:"name" binding:"required"`
	Description string            `json:"description"`
	Schedule    string            `json:"schedule" binding:"required"` // cron expression
	Payload     map[string]string `json:"payload"`
}

// Event represents the response structure
type Event struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Schedule    string            `json:"schedule"`
	Payload     map[string]string `json:"payload"`
	CreatedAt   time.Time         `json:"created_at"`
}

// @Summary Create a new event
// @Description Create a eventbridge the provided data
// @Accept json
// @Produce json
// @Param post body controllers.CreateEventRequest true "Event data"
// @Success 201 {object} controllers.Event
// @Failure 500 {object} object "Internal server error"
// @Router /event [post]
func CreateEvent(c *gin.Context) {
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
		State:             types.RuleStateEnabled,
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
				Id:      aws.String(fmt.Sprintf("%s-target", req.Name)),
				Arn:     aws.String("YOUR_LAMBDA_FUNCTION_ARN"), // Replace with your Lambda function ARN
				Input:   aws.String(string(payloadJSON)),
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
