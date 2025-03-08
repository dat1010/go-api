package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// Cool now lets add the main function

func main() {
	router := gin.Default()
	router.GET("/api/healthcheck", getVersion)

	router.Run("localhost:8080")
}

type health struct {
	Version string `json:"version"`
}

var checkData = []health{{Version: "0.0.1"}}

func getVersion(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, checkData)
}

//Fun stuff, I do not like how its organized :(
//I think ill try to push what we have sofar to AWS ECS now. start small
