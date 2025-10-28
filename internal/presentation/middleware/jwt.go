package middleware

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

const (
	roleAdmin = "admin"
	roleUser  = "user"
	roleGuest = "guest"
)

const (
	keyUserID   string = "userId"
	keyRole     string = "role"
	keyVerified string = "verified"
)

func JWTMiddleware(role string, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {

		if strings.Contains(c.Request.URL.Path, "login") {
			c.Next()
			return
		}

		secretKey := []byte(os.Getenv("HS_256KEY"))
		if len(secretKey) == 0 {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "SERVER_ERROR",
				"message": "Server configuration error",
			})
			c.Abort()
			return
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "UNAUTHORIZED",
				"message": "Authorization header required",
			})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return secretKey, nil
		})

		log.Debug("token", zap.String("service", "JWTMiddleware"), zap.Any("token", token))

		if err != nil || !token.Valid {
			log.Debug("invalid token", zap.String("service", "JWTMiddleware"), zap.Error(err))
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "INVALID_TOKEN",
				"message": "Invalid or expired token",
			})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "INVALID_TOKEN",
				"message": "Invalid token claims",
			})
			c.Abort()
			return
		}

		userRole, ok := claims["role"].(string)
		if !ok || userRole != role {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "FORBIDDEN",
				"message": fmt.Sprintf("Access denied. Required role: %s", role),
			})
			c.Abort()
			return
		}

		// c.Set("user_id", claims["user_id"])
		// c.Set("email", claims["email"])
		// c.Set("role", claims["role"])
		// c.Set("verified", claims["verified"])

		ctx := c.Request.Context()
		ctx = context.WithValue(ctx, keyUserID, claims["user_id"])
		ctx = context.WithValue(ctx, keyRole, claims["role"])
		ctx = context.WithValue(ctx, keyVerified, claims["verified"])
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}
