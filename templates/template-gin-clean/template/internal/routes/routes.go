package routes

import (
	"database/sql"
	"net/http"

	"{{MODULE_NAME}}/internal/config"
	"{{MODULE_NAME}}/internal/middleware"
	"github.com/gin-gonic/gin"
)

// Setup registers all application routes and returns the Gin engine.
func Setup(cfg *config.Config, db *sql.DB) *gin.Engine {
	r := gin.New()
	r.Use(middleware.RequestID())
	r.Use(middleware.Logger())
	r.Use(middleware.Recovery())
	r.Use(middleware.CORS())

	api := r.Group("/api/v1")
	api.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "app": cfg.AppName})
	})
	api.GET("/live", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "alive"})
	})
	api.GET("/ready", func(c *gin.Context) {
		if err := db.Ping(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not ready"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})

	// gostack:routes

	return r
}
