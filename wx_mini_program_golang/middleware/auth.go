package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"
	"wxcloudrun-golang/util"

	"github.com/cloudwego/hertz/pkg/app"
)

// JWTAuth Middleware to validate JWT token
func JWTAuth() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		authHeader := string(c.GetHeader("Authorization"))
		var tokenString string

		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid authorization header format"})
				c.Abort()
				return
			}
			tokenString = parts[1]
		} else {
			// Fallback for scenarios where header cannot be set (e.g., <image src>)
			qToken := c.Query("token")
			if qToken == "" {
				c.JSON(http.StatusUnauthorized, map[string]string{"error": "Authorization header is missing"})
				c.Abort()
				return
			}
			tokenString = qToken
		}

		claims, err := util.ParseToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Set OpenID in context
		c.Set("openid", claims.OpenID)

		// Token Refresh Logic (Sliding Window)
		// If token is valid but expires soon (e.g., within half of its lifetime), issue a new one
		// Total lifetime is 10 min. If remaining time < 5 min, refresh.
		if claims.ExpiresAt != nil {
			remaining := claims.ExpiresAt.Time.Sub(time.Now())
			if remaining > 0 && remaining < util.JWTExpireTime/2 {
				newToken, err := util.GenerateToken(claims.OpenID)
				if err == nil {
					c.Header("New-Token", newToken)
					// Optional: Log refresh action
					// fmt.Printf("Refreshed token for user %s\n", claims.OpenID)
				}
			}
		}

		c.Next(ctx)
	}
}
