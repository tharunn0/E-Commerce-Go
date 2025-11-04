package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/tharunn0/E-Commerce-Go/internal/presentation/handler"
	"github.com/tharunn0/E-Commerce-Go/internal/presentation/middleware"
	"go.uber.org/zap"
)

func RegisterRoutes(g *gin.Engine, logger *zap.Logger, userh *handler.UserHandler, adminh *handler.AdminHandler, categoryh *handler.CategoryHandler, producth *handler.ProductHandler) {

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
		userAuth.POST("/google")
	}
	{
		userProtected := g.Group("/api/v1/users/").Use(middleware.JWTMiddleware("user", logger))
		userProtected.GET("/profile", userh.GetProfile)
	}

	adminAuth := g.Group("/api/v1/auth/admin/")
	adminAuth.POST("/login", adminh.LoginUser)

	adminProtectedRoute := g.Group("/api/v1/admin").Use(middleware.JWTMiddleware("admin", logger))
	adminProtectedRoute.GET("/users", adminh.GetAllUsers)
	adminProtectedRoute.PUT("/users", adminh.UpdateUserStatus)
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

	{
		brandRoute := g.Group("/api/v1/brands")
		// Brand routes
		brandProtectedRoute := brandRoute.Group("/").Use(middleware.JWTMiddleware("admin", logger))
		brandProtectedRoute.POST("/", producth.CreateBrand)
		brandProtectedRoute.PUT("/", producth.UpdateBrand)
		brandProtectedRoute.DELETE("/:id", producth.DeleteBrand)

		brandOpenRoute := brandRoute.Group("/").Use(middleware.AuthContextMiddleware(logger))
		brandOpenRoute.GET("/", producth.GetBrands)
		brandOpenRoute.GET("/:id", producth.GetBrandByID)
	}

	{
		// Product routes
		productRoute := g.Group("/api/v1/products")
		// Product routes
		productOpenRoute := productRoute.Group("/").Use(middleware.AuthContextMiddleware(logger))
		productOpenRoute.GET("/", producth.GetProducts)
		productOpenRoute.GET("/:id", producth.GetProductByID)
		productOpenRoute.GET("/:id/variants", producth.GetVariantsByProductID)
		//productOpenRoute.GET("/:id/attributes", producth.GetProductAttributes)
		//productOpenRoute.GET("/:id/reviews", producth.GetProductReviews)

		productProtectedRoute := productRoute.Group("/").Use(middleware.JWTMiddleware("admin", logger))
		productProtectedRoute.POST("/", producth.CreateProduct)
		productProtectedRoute.PUT("/", producth.UpdateProduct)
		productProtectedRoute.DELETE("/:id", producth.DeleteProduct)
	}

	{
		// Product variant routes
		productVariantRoute := g.Group("/api/v1/product-variants")
		productVariantOpenRoute := productVariantRoute.Group("/").Use(middleware.AuthContextMiddleware(logger))
		productVariantOpenRoute.GET("/:id", producth.GetProductVariantByID)
		// productVariantOpenRoute.GET("/", producth.GetProductVariants)
		// productVariantOpenRoute.GET("/:id", producth.GetProductVariantByID)

		productVariantProtectedRoute := productVariantRoute.Group("/").Use(middleware.JWTMiddleware("admin", logger))
		productVariantProtectedRoute.POST("/", producth.CreateProductVariant)
		// productVariantProtectedRoute.POST("/", producth.CreateProductVariant)
		// productVariantProtectedRoute.PUT("/", producth.UpdateProductVariant)
		productVariantProtectedRoute.DELETE("/:id", producth.DeleteProductVariant)

	}

	{
		// Variant attribute routes
		variantAttributeRoute := g.Group("/api/v1/attributes")
		// Variant attribute routes
		variantAttributeProtectedRoute := variantAttributeRoute.Group("/").Use(middleware.JWTMiddleware("admin", logger))
		variantAttributeProtectedRoute.POST("/", producth.CreateAttribute)
		// variantAttributeProtectedRoute.PUT("/", producth.UpdateVariantAttribute)
		// variantAttributeProtectedRoute.DELETE("/:id", producth.DeleteVariantAttribute)

		variantAttributeOpenRoute := variantAttributeRoute.Group("/").Use(middleware.AuthContextMiddleware(logger))
		variantAttributeOpenRoute.GET("/", producth.GetAttributes)
		variantAttributeOpenRoute.GET("/:id", producth.GetAttributeByID)
	}

}
