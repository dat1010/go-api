package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

const discordWebhookURL = "https://discord.com/api/webhooks/1463686090308456660/a_OjI80L-0Od2CPIEll_ECnvsGkENtBeDCAOEuBBkEA3C1nkpqyYWGu73lEIPpZGVtfR"

type discordWebhookPayload struct {
	Content string `json:"content"`
}

// DiscordPingResponse represents a successful ping response.
type DiscordPingResponse struct {
	Message       string `json:"message"`
	DiscordStatus string `json:"discord_status"`
}

// PingDiscord godoc
// @Summary      Ping Discord webhook
// @Description  Send a test notification to Discord for user tannerd
// @Tags         notifications
// @Produce      json
// @Success      200  {object}  DiscordPingResponse
// @Failure      500  {object}  map[string]string
// @Router       /api/discord-ping [get]
func PingDiscord(c *gin.Context) {
	message := fmt.Sprintf(
		"Someone is pinging you, they have questions at your presentation table. @tannerd (sent at %s)",
		time.Now().UTC().Format(time.RFC3339),
	)

	payload := discordWebhookPayload{Content: message}
	body, err := json.Marshal(payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to marshal Discord payload"})
		return
	}

	req, err := http.NewRequestWithContext(c.Request.Context(), http.MethodPost, discordWebhookURL, bytes.NewBuffer(body))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create Discord request"})
		return
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to send Discord webhook: %v", err)})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Discord webhook returned status %s", resp.Status)})
		return
	}

	c.JSON(http.StatusOK, DiscordPingResponse{
		Message:       "Discord ping sent for user tannerd",
		DiscordStatus: resp.Status,
	})
}
