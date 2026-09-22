package controller

import (
	"context"
	"net/http"
	"one-api/model"
	"time"

	"github.com/gin-gonic/gin"
)

// Health checks the process and its database without exposing configuration.
func Health(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()
	if model.DB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"ok": false})
		return
	}
	db, err := model.DB.DB()
	if err != nil || db.PingContext(ctx) != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"ok": false})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
