package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/tharunn0/E-Commerce-Go/internal/config"
	"github.com/tharunn0/E-Commerce-Go/internal/presentation/handler"
	"github.com/tharunn0/E-Commerce-Go/internal/presentation/middleware"
	"go.uber.org/zap"
)

type Handler struct {
	User     *handler.UserHandler
	Admin    *handler.AdminHandler
	Category *handler.CategoryHandler
	Product  *handler.ProductHandler
}

func NewHandler(userh *handler.UserHandler, adminh *handler.AdminHandler, categoryh *handler.CategoryHandler, producth *handler.ProductHandler) *Handler {
	return &Handler{
		User:     userh,
		Admin:    adminh,
		Category: categoryh,
		Product:  producth,
	}
}

func RegisterRoutes(g *gin.Engine, logger *zap.Logger, h *Handler, cfg *config.SecuritySettings) {

	g.GET("/home", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"msg": "Server up and ready to roll",
		})
	})

	{
		userAuth := g.Group("/api/v1/auth/users/")
		userAuth.POST("/register", h.User.RegisterUser)
		userAuth.POST("/login", h.User.LoginUser)
		userAuth.POST("/reset-password-link", h.User.SendPasswordResetLink)
		userAuth.POST("/reset-password/", h.User.ResetPassword)
		userAuth.GET("/google", h.User.GoogleSignIn)
		userAuth.GET("/google/callback", h.User.GoogleCallback)
	}
	{
		userProtected := g.Group("/api/v1/users/").Use(middleware.JWTMiddleware("user", logger, cfg.JWTSecret))
		userProtected.GET("/profile", h.User.GetProfile)
		userProtected.POST("/verify-otp", h.User.VerifyOTP)
		userProtected.POST("/send-otp", h.User.SendOTP)
	}

	adminAuth := g.Group("/api/v1/auth/admin/")
	adminAuth.POST("/login", h.Admin.LoginUser)

	adminProtectedRoute := g.Group("/api/v1/admin").Use(middleware.JWTMiddleware("admin", logger, cfg.JWTSecret))
	adminProtectedRoute.GET("/users", h.Admin.GetAllUsers)
	adminProtectedRoute.PUT("/users", h.Admin.UpdateUserStatus)
	{
		// Category routes
		categoryOpenRoute := g.Group("/api/v1/categories/").Use(middleware.AuthContextMiddleware(logger, cfg.JWTSecret))
		categoryOpenRoute.GET("/", h.Category.GetAllCategories)
		categoryOpenRoute.GET("/:id", h.Category.GetCategoryByID)

		categoryProtectedRoute := g.Group("/api/v1/categories/").Use(middleware.JWTMiddleware("admin", logger, cfg.JWTSecret))
		categoryProtectedRoute.POST("/", h.Category.CreateCategory)
		categoryProtectedRoute.DELETE("/:id", h.Category.DeleteCategory)
		categoryProtectedRoute.PUT("/", h.Category.UpdateCategory)
		categoryProtectedRoute.PATCH("/:id/activate", h.Category.ToggleCategoryStatus)
		categoryProtectedRoute.PATCH("/:id/deactivate", h.Category.ToggleCategoryStatus)

	}

	{
		brandRoute := g.Group("/api/v1/brands")
		// Brand routes
		brandProtectedRoute := brandRoute.Group("/").Use(middleware.JWTMiddleware("admin", logger, cfg.JWTSecret))
		brandProtectedRoute.POST("/", h.Product.CreateBrand)
		brandProtectedRoute.PUT("/", h.Product.UpdateBrand)
		brandProtectedRoute.DELETE("/:id", h.Product.DeleteBrand)

		brandOpenRoute := brandRoute.Group("/").Use(middleware.AuthContextMiddleware(logger, cfg.JWTSecret))
		brandOpenRoute.GET("/", h.Product.GetBrands)
		brandOpenRoute.GET("/:id", h.Product.GetBrandByID)
	}

	{
		// Product routes
		productRoute := g.Group("/api/v1/products")
		// Product routes
		productOpenRoute := productRoute.Group("/").Use(middleware.AuthContextMiddleware(logger, cfg.JWTSecret))
		productOpenRoute.GET("/", h.Product.GetProducts)
		productOpenRoute.GET("/:id", h.Product.GetProductByID)
		productOpenRoute.GET("/:id/variants", h.Product.GetVariantsByProductID)
		//productOpenRoute.GET("/:id/attributes", h.Product.GetProductAttributes)
		//productOpenRoute.GET("/:id/reviews", h.Product.GetProductReviews)

		productProtectedRoute := productRoute.Group("/").Use(middleware.JWTMiddleware("admin", logger, cfg.JWTSecret))
		productProtectedRoute.POST("/", h.Product.CreateProduct)
		productProtectedRoute.PUT("/", h.Product.UpdateProduct)
		productProtectedRoute.DELETE("/:id", h.Product.DeleteProduct)
	}

	{
		// Product variant routes
		productVariantRoute := g.Group("/api/v1/product-variants")
		productVariantOpenRoute := productVariantRoute.Group("/").Use(middleware.AuthContextMiddleware(logger, cfg.JWTSecret))
		productVariantOpenRoute.GET("/:id", h.Product.GetProductVariantByID)
		// productVariantOpenRoute.GET("/", h.Product.GetProductVariants)
		// productVariantOpenRoute.GET("/:id", h.Product.GetProductVariantByID)

		productVariantProtectedRoute := productVariantRoute.Group("/").Use(middleware.JWTMiddleware("admin", logger, cfg.JWTSecret))
		productVariantProtectedRoute.POST("/", h.Product.CreateProductVariant)
		// productVariantProtectedRoute.POST("/", h.Product.CreateProductVariant)
		// productVariantProtectedRoute.PUT("/", h.Product.UpdateProductVariant)
		productVariantProtectedRoute.DELETE("/:id", h.Product.DeleteProductVariant)

	}

	{
		// Variant attribute routes
		variantAttributeRoute := g.Group("/api/v1/attributes")
		// Variant attribute routes
		variantAttributeProtectedRoute := variantAttributeRoute.Group("/").Use(middleware.JWTMiddleware("admin", logger, cfg.JWTSecret))
		variantAttributeProtectedRoute.POST("/", h.Product.CreateAttribute)
		// variantAttributeProtectedRoute.PUT("/", h.Product.UpdateVariantAttribute)
		// variantAttributeProtectedRoute.DELETE("/:id", h.Product.DeleteVariantAttribute)

		variantAttributeOpenRoute := variantAttributeRoute.Group("/").Use(middleware.AuthContextMiddleware(logger, cfg.JWTSecret))
		variantAttributeOpenRoute.GET("/", h.Product.GetAttributes)
		variantAttributeOpenRoute.GET("/:id", h.Product.GetAttributeByID)
	}

}
