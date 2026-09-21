package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	jwtpkg "github.com/thriftin/api/pkg/jwt"
	"github.com/thriftin/api/pkg/response"
)

func Auth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.GetHeader("Authorization")
		if !strings.HasPrefix(h, "Bearer ") {
			response.Err(c, 401, "unauthorized"); c.Abort(); return
		}
		claims, err := jwtpkg.Validate(secret, strings.TrimPrefix(h, "Bearer "))
		if err != nil { response.Err(c, 401, "invalid token"); c.Abort(); return }
		c.Set("userID", claims.UserID)
		c.Set("role", claims.Role)
		c.Next()
	}
}
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		r, _ := c.Get("role")
		for _, a := range roles { if r == a { c.Next(); return } }
		response.Err(c, 403, "forbidden"); c.Abort()
	}
}
