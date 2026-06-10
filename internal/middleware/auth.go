package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/qq900306ss/linda-salon-api/internal/auth"
)

// Context keys set by the auth middleware.
const (
	UsernameKey = "auth_username"
	RoleKey     = "auth_role"
)

func abortUnauthorized(c *gin.Context, code, message string) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
		"success": false,
		"error":   gin.H{"code": code, "message": message},
	})
}

// AdminRequired validates the Bearer JWT and requires the admin role.
func AdminRequired(jwtManager *auth.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			abortUnauthorized(c, "UNAUTHORIZED", "Authorization token required")
			return
		}
		token := strings.TrimPrefix(header, "Bearer ")

		claims, err := jwtManager.ValidateToken(token)
		if err != nil {
			abortUnauthorized(c, "UNAUTHORIZED", "Invalid or expired token")
			return
		}
		if claims.Role != "admin" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   gin.H{"code": "FORBIDDEN", "message": "Admin access required"},
			})
			return
		}

		c.Set(UsernameKey, claims.Username)
		c.Set(RoleKey, claims.Role)
		c.Next()
	}
}
