package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const discordWebhookURL = ""

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
	html := `<!doctype html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>Ping Sent</title>
  <style>
    body { font-family: Arial, sans-serif; background: #f7f7f7; color: #222; padding: 40px; }
    .card { max-width: 420px; margin: 0 auto; background: #fff; border-radius: 8px; padding: 24px; box-shadow: 0 4px 12px rgba(0,0,0,0.08); }
    h1 { margin-top: 0; font-size: 22px; }
    p { margin: 12px 0 0; line-height: 1.5; }
  </style>
</head>
<body>
  <div class="card">
    <h1>Ping Sent</h1>
    <p>This feature is no longer available. Please use the <a href="https://github.com/tannerdolby/go-api">go-api</a> instead.</p>
  </div>
</body>
</html>`
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
}
