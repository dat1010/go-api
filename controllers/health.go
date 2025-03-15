package controllers

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type Health struct {
	Version string `json:"version"`
}

var healthCheckData = Health{Version: "0.0.7"}

func GetHealthCheck(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, healthCheckData)
}
