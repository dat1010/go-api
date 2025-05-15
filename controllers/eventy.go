package controllers

import "github.com/gin-gonic/gin"

// @Summary Create a new event
// @Description Create a eventbridge the provided data
// @Accept json
// @Produce json
// @Param post body controllers.CreateEventRequest true "EvenPostt data"
// @Success 201 {object} controllers.Event
// @Failure 500 {object} object "Internal server error"
// @Router /event [post]

func CreateEvent(c *gin.Context) {
}
