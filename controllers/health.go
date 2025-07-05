package controllers

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

type Health struct {
	Version string `json:"version"`
}

var (
	version         = os.Getenv("VERSION")
	healthCheckData = Health{Version: version}
)

// GetHealthCheck godoc
// @Summary      Health Check
// @Description  Return service version
// @Tags         health
// @Produce      json
// @Success      200  {object}  Health
// @Router       /api/healthcheck [get]
func GetHealthCheck(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, healthCheckData)
}
