package controllers

import (
	"net/http"
	"os"
	"runtime"

	"github.com/gin-gonic/gin"
)

type Health struct {
	Version string `json:"version"`
}

type GoVersion struct {
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
// @Router       /healthcheck [get]
func GetHealthCheck(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, healthCheckData)
}

// GetGoVersion godoc
// @Summary      Go Version
// @Description  Return installed Go runtime version
// @Tags         health
// @Produce      json
// @Success      200  {object}  GoVersion
// @Router       /version [get]
func GetGoVersion(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, GoVersion{Version: runtime.Version()})
}
