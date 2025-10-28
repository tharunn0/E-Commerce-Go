package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/tharunn0/E-Commerce-Go/internal/presentation/handler"
	"github.com/tharunn0/E-Commerce-Go/internal/presentation/middleware"
	"go.uber.org/zap"
)

func RegisterRoutes(g *gin.Engine, logger *zap.Logger, userh *handler.UserHandler) {

	g.GET("/home", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"msg": "server up and ready to roll",
		})
	})

	userAuth := g.Group("/api/v1/auth/users/")
	{
		userAuth.POST("/register", userh.RegisterUser)
		userAuth.POST("/login", userh.LoginUser)
		userAuth.POST("/send-otp", userh.SendOTP)
		userAuth.POST("/verify-otp", userh.VerifyOTP)
		userAuth.POST("/reset-password-link", userh.SendPasswordResetLink)
		userAuth.POST("/reset-password/", userh.ResetPassword)
	}
	userProtected := g.Group("/api/v1/users/")
	userProtected.Use(middleware.JWTMiddleware("user", logger))
	{
		userProtected.GET("/profile", userh.GetProfile)
	}

}
