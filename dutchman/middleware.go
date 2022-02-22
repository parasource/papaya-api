package dutchman

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

func (d *Dutchman) AuthMiddleware(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	headerParts := strings.Split(authHeader, " ")
	if len(headerParts) != 2 {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	if headerParts[0] != "Bearer" {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	_, err := ParseToken(headerParts[1])
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
}
