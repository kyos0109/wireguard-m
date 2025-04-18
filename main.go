package main

import (
	"embed"
	"flag"
	"html/template"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"

	"github.com/kyos0109/WireGuard-M/models"
	"github.com/kyos0109/WireGuard-M/utils"
)

//go:embed templates/*
var templatesFS embed.FS

func loadTemplates() *template.Template {
	return template.Must(template.ParseFS(templatesFS, "templates/*.html"))
}

func main() {
	r := gin.Default()
	r.SetTrustedProxies([]string{"127.0.0.1"})

	models.PeersStorePath = flag.String("peers", "data/peers.json", "Peers Store Path")
	models.ConfigPath = flag.String("config", "data/config.json", "Config Path")
	flag.Parse()

	r.SetHTMLTemplate(loadTemplates())

	routerEntry(r)

	go func() {
		if err := r.Run(":8080"); err != nil {
			log.Fatalf("Gin server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	log.Println("Shutting down server...")

	utils.CloseWGClient()

	log.Println("Server exited gracefully.")
}
