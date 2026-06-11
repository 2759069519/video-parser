package api

import (
	"video-parser/internal/handler"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	videoHandler := handler.NewVideoHandler()

	api := r.Group("/api")
	{
		api.POST("/parse", videoHandler.Parse)
		api.POST("/fetch-video-url", videoHandler.FetchVideoURL)
		api.POST("/fetch-atlas-images", videoHandler.FetchAtlasImages)
		api.GET("/download", videoHandler.Download)
		api.GET("/proxy-image", videoHandler.ProxyImage)

		api.GET("/video/:id", videoHandler.GetVideo)
		api.GET("/atlas/:id", videoHandler.GetAtlas)
		api.GET("/profile/:id", videoHandler.GetProfile)

		api.GET("/records", videoHandler.GetRecords)
	}
}
