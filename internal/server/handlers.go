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

// GetVMByDomainIDReq describes request needed to get VM by its ID.
type GetVMByDomainIDReq struct {
	// ID is the only requirement for fetching a VM.
	DomainID int `uri:"domain_id" binding:"required"`
}

func (a *App) getVMByDomainID(c *gin.Context) {
	var req GetVMByDomainIDReq
	if err := c.ShouldBindUri(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	vm, err := a.manager.GetVMByDomainID(req.DomainID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": gin.H{"vm": vm}})
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

	vm, err := a.manager.CreateVM(vmReq.MachineType, vmReq.Memory, vmReq.VCPUs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if _, err := a.storage.CreateVM(vm); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "success"})
}

// TakeMemorySnapshots gets the currently used percentage of total memory for each VM and
// persists the measurements to our postgres storage... This will be used to fetch data for
// charts on the front end.
func (a *App) TakeMemorySnapshots() {
	vms, err := a.manager.GetRunningVMs()
	if err != nil {
		a.logger.Errorf("failed to get running vms, err: %+v", err)
	}

	for _, vm := range vms {
		usage := (float64(vm.Mem-vm.CurrentMem) / float64(vm.Mem)) * 100
		if err := a.storage.RecordVMMemory(vm.DomainID, usage); err != nil {
			a.logger.Errorf("failed to record VM memory, err: %+v", err)
		}
	}
}
