package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/kyos0109/WireGuard-M/controllers"
	"github.com/kyos0109/WireGuard-M/models"
)

func routerEntry(r *gin.Engine) {
	_, err := controllers.EnsureWgmPasswdFile()
	if err != nil {
		log.Fatalf("Error ensuring password file: %v", err)
	}

	peerStore := models.NewPeerStore()

	r.GET("/", controllers.ShowLogin)
	r.POST("/login", controllers.DoLogin)

	authorized := r.Group("/")
	authorized.Use(controllers.AuthRequired())
	{
		authorized.GET("/dashboard", controllers.ShowDeviceDashboard)
		authorized.GET("/interfaces", controllers.ListDevice)
		authorized.GET("/interfaces/:interfaceName/peers", controllers.GetPeers(peerStore))
		authorized.GET("/peer/add_page", controllers.AddPeerPage)
		authorized.POST("/peer/add", controllers.AddPeer(peerStore))
		authorized.POST("/peer/delete", controllers.DeletePeer(peerStore))
		authorized.GET("/peer/qrcode/:id", controllers.GeneratePeerQR(peerStore))
		authorized.GET("/peer/download_config/:id", controllers.DownloadConfig(peerStore))
	}
}
