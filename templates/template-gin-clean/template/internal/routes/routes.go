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
	r := gin.Default()
	r.Use(middleware.CORS())

	api := r.Group("/api/v1")
	api.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "app": cfg.AppName})
	})

	// gostack:routes

	return r
}
