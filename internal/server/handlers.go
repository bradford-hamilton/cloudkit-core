package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (a *App) ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "pong"})
}

func (a *App) getVMs(c *gin.Context) {
	domains, err := a.manager.GetRunningVMs()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"vms": domains}})
}

// CreateVMReq defines the shape of the JSON request needed from the front end to create a VM.
type CreateVMReq struct {
	// MachineType currently is a placeholder until we have more images than ubuntu-18.04
	MachineType string `json:"machineType" binding:"required"`
	// Memory in this context is an int representing GB
	Memory int `json:"memory" binding:"required"`
	// VCPUs refers to the requested number of vCPUs
	VCPUs int `json:"vcpus" binding:"required"`
}

func (a *App) createVM(c *gin.Context) {
	var vmReq CreateVMReq
	if err := c.ShouldBindJSON(&vmReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: add metadata to database
	err := a.manager.CreateVM(vmReq.MachineType, vmReq.Memory, vmReq.VCPUs)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "success"})
}
