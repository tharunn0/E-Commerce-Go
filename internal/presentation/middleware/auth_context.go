package middleware

import (
	"context"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/tharunn0/E-Commerce-Go/internal/domain"
	"go.uber.org/zap"
)

func AuthContextMiddleware(log *zap.Logger, jwtsecret string) gin.HandlerFunc {
	if len(jwtsecret) == 0 {
		log.Fatal("JWT secret is required")
	}
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		if strings.Contains(path, "/login") || strings.Contains(path, "/register") {
			ctx := c.Request.Context()
			ctx = context.WithValue(ctx, domain.KeyUserID, "")
			ctx = context.WithValue(ctx, domain.KeyRole, roleGuest)
			ctx = context.WithValue(ctx, domain.KeyVerified, false)
			c.Request = c.Request.WithContext(ctx)
			c.Next()
			return
		}

		role := roleGuest
		userID := 0
		verified := false

		auth := c.GetHeader("Authorization")
		if strings.HasPrefix(auth, "Bearer ") {
			tokenStr := strings.TrimPrefix(auth, "Bearer ")

			token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(jwtsecret), nil
			})

			if err == nil && token.Valid {
				if claims, ok := token.Claims.(jwt.MapClaims); ok {
					if r, ok := claims["role"].(string); ok && (r == roleAdmin || r == roleUser) {
						role = r
					}
					if uid, ok := claims["user_id"].(float64); ok {
						userID = int(uid)
					}
					verified = claims["verified"] == true
					log.Debug("JWT validated", zap.Int("userId", userID), zap.String("role", role))
				}
			} else {
				log.Debug("Invalid JWT", zap.Error(err))
			}
		} else {
			log.Debug("No Bearer token : Continuing as guest")
		}

		ctx := c.Request.Context()
		ctx = context.WithValue(ctx, domain.KeyUserID, userID)
		ctx = context.WithValue(ctx, domain.KeyRole, role)
		ctx = context.WithValue(ctx, domain.KeyVerified, verified)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}
