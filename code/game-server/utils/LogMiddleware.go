package utils

import (
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func JSONLogMiddleware(context *gin.Context) {
	startTime := time.Now()
	context.Next()
	duration := time.Since(startTime).Milliseconds()

	entry := log.WithFields(log.Fields{
		"duration": duration,
		"method":   context.Request.Method,
		"path":     context.Request.RequestURI,
		"status":   context.Writer.Status(),
		"referrer": context.Request.Referer(),
	})

	if context.Writer.Status() >= 500 {
		entry.Error(context.Errors.String())
	} else {
		entry.Info("")
	}
}
