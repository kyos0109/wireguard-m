package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/kyos0109/WireGuard-M/utils"
)

func ShowDeviceDashboard(c *gin.Context) {
	c.HTML(http.StatusOK, "dashboard.html", nil)
}

type WireGuardInterface struct {
	Name string `json:"name"`
}

func ListDevice(c *gin.Context) {
	ds, err := utils.ListDevices()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ds)
}
