package server

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (s Server) middlewareCallID(c *gin.Context) error {
	start := time.Now()
	if len(c.Request.Header.Get("X-Call-Id")) == 0 {
		callID := uuid.New().String()
		c.Request.Header.Add("X-Call-Id", callID)
		c.Writer.Header().Add("X-Call-Id", callID)
	}

	c.Request.Header.Add("X-Start-Time", fmt.Sprintf("%d", start.UTC().UnixNano()))
	c.Writer.Header().Add("X-Start-Time", fmt.Sprintf("%d", start.UTC().UnixNano()))

	c.Next()
	return nil
}
