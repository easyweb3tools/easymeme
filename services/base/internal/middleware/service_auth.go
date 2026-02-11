package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func ServiceAuthMiddleware(serviceTokens map[string]string) gin.HandlerFunc {
	if len(serviceTokens) == 0 {
		return func(c *gin.Context) {
			c.Set("service_id", "anonymous")
			c.Next()
		}
	}

	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			return
		}
		token := strings.TrimSpace(strings.TrimPrefix(auth, "Bearer "))
		if token == "" || token == auth {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization format"})
			return
		}
		for serviceID, expected := range serviceTokens {
			if token == expected {
				c.Set("service_id", serviceID)
				c.Next()
				return
			}
		}
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid service token"})
	}
}
