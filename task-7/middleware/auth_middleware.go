package middleware

import (
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const (
	CtxUserIDKey   = "user_id"
	CtxUsernameKey = "username"
	CtxRoleKey     = "role"
)

func AuthRequired() gin.HandlerFunc {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		panic("JWT_SECRET is not set")
	}
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if !strings.HasPrefix(strings.ToLower(auth), "bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid Authorization header"})
			return
		}
		tokenString := strings.TrimSpace(auth[len("Bearer "):])

		claims := jwt.MapClaims{}
		tkn, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		}, jwt.WithLeeway(5*time.Second))
		if err != nil || !tkn.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		uid, _ := claims["uid"].(string)
		username, _ := claims["username"].(string)
		role, _ := claims["role"].(string)

		c.Set(CtxUserIDKey, uid)
		c.Set(CtxUsernameKey, username)
		c.Set(CtxRoleKey, role)
		c.Next()
	}
}

func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get(CtxRoleKey)
		if role != "admin" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "admin access required"})
			return
		}
		c.Next()
	}
}
