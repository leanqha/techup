package account

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware(service ServiceInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		accessToken, _ := c.Cookie("access_token")
		var claims jwt.MapClaims
		var err error

		if accessToken != "" {
			claims, err = ParseToken(accessToken)
		}

		if err != nil || accessToken == "" {
			refreshToken, _ := c.Cookie("refresh_token")
			if refreshToken != "" {
				newAccess, _, err := service.RefreshTokens(c.Request.Context(), refreshToken)
				if err == nil {
					claims, _ = ParseToken(newAccess)
				}
			}
		}

		if claims == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		c.Set("user_id", int(claims["user_id"].(float64)))
		c.Set("role", claims["role"].(string))
		c.Set("claims", claims)

		c.Next()
	}
}

func RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, exists := c.Get("claims")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing claims"})
			return
		}

		userClaims, ok := claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid claims"})
			return
		}

		userRole, ok := userClaims["role"].(string)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid role"})
			return
		}

		if userRole != role {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}

		c.Next()
	}
}
