package controllers

import (
	"net/http"

 	"github.com/gin-gonic/gin"
)

// HelloWorld returns a plain "hello" string
func HelloWorld(c *gin.Context) {
 	c.String(http.StatusOK, "hello")
}
