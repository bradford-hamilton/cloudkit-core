package server

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (a *App) ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "pong"})
}

// TODO: implement
func (a *App) getVMs(c *gin.Context) {
	domains, err := a.manager.GetVMs()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"vms": domains}})
}

// CreateVMReq ...
type CreateVMReq struct {
	MachineType string `json:"machineType" binding:"required"`
}

func (a *App) createVM(c *gin.Context) {
	var vmReq CreateVMReq
	if err := c.ShouldBindJSON(&vmReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO:
	// - add metadata to database
	if err := a.manager.CreateVM(); err != nil {
		fmt.Println(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "success"})
}
