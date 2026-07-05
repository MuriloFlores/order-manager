package server

import (
	"os"
	"strconv"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type RoutesInterface interface {
	RegisterRoutes(router *gin.RouterGroup)
}

func RegisterRouter(routes ...RoutesInterface) *gin.Engine {
	r := gin.New()

	r.Use(gin.Recovery())
	r.Use(configureCORS())

	api := r.Group("/api/v1")

	api.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	setupPublicRoutes(api, routes...)

	return r
}

func setupPublicRoutes(r *gin.RouterGroup, routes ...RoutesInterface) {
	publicGroup := r.Group("/")

	for _, route := range routes {
		route.RegisterRoutes(publicGroup)
	}
}

func configureCORS() gin.HandlerFunc {
	isProdStr := os.Getenv("IS_PROD")
	isProd, err := strconv.ParseBool(isProdStr)
	if err != nil {
		isProd = false
	}

	config := cors.DefaultConfig()
	config.AllowAllOrigins = !isProd
	config.AllowHeaders = []string{
		"Authorization",
		"Content-Type",
		"Origin",
	}

	config.ExposeHeaders = []string{"Content-Length"}

	return cors.New(config)
}
