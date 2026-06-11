package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"video-parser/api"
	"video-parser/internal/config"
	"video-parser/internal/repository"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("配置加载失败:", err)
	}
	repository.InitDB(&cfg.Database)

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	api.SetupRoutes(r)

	frontendDir := "./frontend/dist"
	r.Static("/assets", filepath.Join(frontendDir, "assets"))
	r.NoRoute(func(c *gin.Context) {
		indexPath := filepath.Join(frontendDir, "index.html")
		if _, err := os.Stat(indexPath); err == nil {
			c.File(indexPath)
		} else {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		}
	})

	log.Printf("服务器启动在端口 %s", cfg.Server.Port)
	if err := r.Run(":" + cfg.Server.Port); err != nil {
		log.Fatal("服务器启动失败:", err)
	}
}
