package server

import (
	"net/http"

	"github.com/bradford-hamilton/cloudkit-core/internal/cloudkit"
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
	// DomainID is the only requirement for fetching a VM.
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

	id, err := a.storage.GetVMIDFromDomainID(vm.DomainID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	usages, err := a.storage.GetLast15MinVMMemUsage(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": gin.H{"vm": vm, "memory_usage": usages}})
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

// takeMemorySnapshots gets the currently used percentage of total memory for each VM and
// persists the measurements to storage (postgres).
func (a *App) takeMemorySnapshots() {
	vms, err := a.manager.GetRunningDomains()
	if err != nil {
		a.logger.Errorf("failed to get running vms, err: %+v", err)
	}

	for _, domain := range vms {
		rStats, err := a.manager.DomainMemoryStats(domain, cloudkit.MaxStats, 0)
		if err != nil {
			a.logger.Errorf("failed to aqcuire domain memory stats, err: %+v", err)
		}

		ms, err := cloudkit.NewMemStats(rStats)
		if err != nil {
			a.logger.Errorf("failed to unmarshal memory stats, err: %+v", err)
		}

		usage := (float64(ms.Available-ms.Usable) / float64(ms.Available)) * 100

		err = a.storage.RecordVMMemory(int(domain.ID), usage)
		if err != nil {
			a.logger.Errorf("failed to record VM memory, err: %+v", err)
		}
	}
}
