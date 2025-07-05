package controllers

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

type Secret struct {
	Value string `json:"secret"`
}

var (
	myLittleSecret = os.Getenv("MY_LITTLE_SECRET")
	viewSecrets    = Secret{Value: myLittleSecret}
)

func GetSecret(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, viewSecrets)
}
