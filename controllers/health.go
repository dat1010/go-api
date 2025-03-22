package controllers

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
)

type Health struct {
	Version string `json:"version"`
}

var version = os.Getenv("VERSION")
var healthCheckData = Health{Version: version}

func GetHealthCheck(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, healthCheckData)
}
