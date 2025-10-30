package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/tharunn0/E-Commerce-Go/internal/presentation/handler"
	"github.com/tharunn0/E-Commerce-Go/internal/presentation/middleware"
	"go.uber.org/zap"
)

func RegisterRoutes(g *gin.Engine, logger *zap.Logger, userh *handler.UserHandler, adminh *handler.AdminHandler, categoryh *handler.CategoryHandler) {

	g.GET("/home", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"msg": "server up and ready to roll",
		})
	})

	{
		userAuth := g.Group("/api/v1/auth/users/")
		userAuth.POST("/register", userh.RegisterUser)
		userAuth.POST("/login", userh.LoginUser)
		userAuth.POST("/send-otp", userh.SendOTP)
		userAuth.POST("/verify-otp", userh.VerifyOTP)
		userAuth.POST("/reset-password-link", userh.SendPasswordResetLink)
		userAuth.POST("/reset-password/", userh.ResetPassword)
	}
	{
		userProtected := g.Group("/api/v1/users/").Use(middleware.JWTMiddleware("user", logger))
		userProtected.GET("/profile", userh.GetProfile)
	}

	adminAuth := g.Group("/api/v1/auth/admin/").Use(middleware.JWTMiddleware("admin", logger))
	adminAuth.POST("/login", adminh.LoginUser)

	{
		// Category routes
		categoryOpenRoute := g.Group("/api/v1/categories/").Use(middleware.AuthContextMiddleware(logger))
		categoryOpenRoute.GET("/", categoryh.GetAllCategories)
		categoryOpenRoute.GET("/:id", categoryh.GetCategoryByID)

		categoryProtectedRoute := g.Group("/api/v1/categories/").Use(middleware.JWTMiddleware("admin", logger))
		categoryProtectedRoute.POST("/", categoryh.CreateCategory)
		categoryProtectedRoute.DELETE("/:id", categoryh.DeleteCategory)
		categoryProtectedRoute.PUT("/", categoryh.UpdateCategory)
		categoryProtectedRoute.PATCH("/:id/activate", categoryh.ToggleCategoryStatus)
		categoryProtectedRoute.PATCH("/:id/deactivate", categoryh.ToggleCategoryStatus)

	}

}
