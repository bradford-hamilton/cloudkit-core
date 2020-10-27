package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (a *API) ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "pong"})
}
