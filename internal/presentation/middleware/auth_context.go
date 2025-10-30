package middleware

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

func AuthContextMiddleware(log *zap.Logger) gin.HandlerFunc {
	secret := []byte(os.Getenv("HS_256KEY"))
	if len(secret) == 0 {
		log.Fatal("HS_256KEY environment variable is required")
	}

	return func(c *gin.Context) {
		path := c.Request.URL.Path
		if strings.Contains(path, "/login") || strings.Contains(path, "/register") {
			ctx := c.Request.Context()
			ctx = context.WithValue(ctx, keyUserID, "")
			ctx = context.WithValue(ctx, keyRole, roleGuest)
			ctx = context.WithValue(ctx, keyVerified, false)
			c.Request = c.Request.WithContext(ctx)
			c.Next()
			return
		}

		role := roleGuest
		userID := ""
		verified := false

		auth := c.GetHeader("Authorization")
		if strings.HasPrefix(auth, "Bearer ") {
			tokenStr := strings.TrimPrefix(auth, "Bearer ")

			token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return secret, nil
			})

			if err == nil && token.Valid {
				if claims, ok := token.Claims.(jwt.MapClaims); ok {
					if r, ok := claims["role"].(string); ok && (r == roleAdmin || r == roleUser) {
						role = r
					}
					if uid, ok := claims["user_id"].(string); ok {
						userID = uid
					}
					verified = claims["verified"] == true
					log.Debug("JWT validated", zap.String("userId", userID), zap.String("role", role))
				}
			} else {
				log.Debug("Invalid JWT", zap.Error(err))
			}
		} else {
			log.Debug("No Bearer token")
		}

		fmt.Println("Role from auth context middleware : ", role)

		ctx := c.Request.Context()
		ctx = context.WithValue(ctx, keyUserID, userID)
		ctx = context.WithValue(ctx, keyRole, role)
		ctx = context.WithValue(ctx, keyVerified, verified)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}
