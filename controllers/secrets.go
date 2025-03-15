package controllers

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
)

type Secret struct {
	Value string `json:"secret"`
}

var myLittleSecret = os.Getenv("MY_LITTLE_SECRET")
var viewSecrets = Secret{Value: myLittleSecret}

func GetSecret(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, viewSecrets)
}
