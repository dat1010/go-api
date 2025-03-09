package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// Cool now lets add the main function

func main() {
	router := gin.Default()
	router.GET("/api/healthcheck", getVersion)

	router.Run("0.0.0.0:8080")
}

type health struct {
	Version string `json:"version"`
}

var checkData = []health{{Version: "0.0.3"}}

func getVersion(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, checkData)
}

// lets open up chatpgt and have it create our github actions to deploy this
// lets see what we get and work our way towards a ci/cd pipeline
// Ill have to go off screen and setup the aws secrets when we get there
// Overall looks like it will work (hopefuly), Lets create the action.
// it will fail at first because i don't have secrets setup
// lets push to main first then break off a release_0.0.2 once we have the
// secrets setup
// of course never push to main!
