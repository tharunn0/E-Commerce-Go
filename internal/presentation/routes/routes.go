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
	Cart     *handler.CartHandler
	Wishlist *handler.WishlistHandler
	Order    *handler.OrderHandler
}

func NewHandler(userh *handler.UserHandler, adminh *handler.AdminHandler, categoryh *handler.CategoryHandler,
	producth *handler.ProductHandler, carth *handler.CartHandler, wishlisth *handler.WishlistHandler, orderh *handler.OrderHandler) *Handler {
	return &Handler{
		User:     userh,
		Admin:    adminh,
		Category: categoryh,
		Product:  producth,
		Cart:     carth,
		Wishlist: wishlisth,
		Order:    orderh,
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
		userAuth.POST("/logout", h.User.LogoutUser)
		userAuth.POST("/refresh", h.User.RefreshToken)
		userAuth.POST("/reset-password-link", h.User.SendPasswordResetLink)
		userAuth.POST("/reset-password/", h.User.ResetPassword)
		userAuth.GET("/google", h.User.GoogleSignIn)
		userAuth.GET("/google/callback", h.User.GoogleCallback)
		userAuth.POST("/verify-otp", h.User.VerifyOTP)
		userAuth.POST("/send-otp", h.User.SendOTP)
	}

	{
		userProtected := g.Group("/api/v1/users/").Use(middleware.JWTMiddleware("user", logger, cfg.JWTSecret))
		userProtected.POST("/email", h.User.EmailChangeRequest)
		userProtected.GET("/email/verify-reset", h.User.VerifyEmailChangeRequest)
		userProtected.GET("/profile", h.User.GetProfile)
		userProtected.PATCH("/profile", h.User.UpdateUserProfile)

		//user address routes
		userAddressProtected := g.Group("/api/v1/users/").Use(middleware.JWTMiddleware("user", logger, cfg.JWTSecret))
		userAddressProtected.POST("/addresses", h.User.CreateUserAddress)
		userAddressProtected.GET("/:id/addresses", h.User.GetUserAddresses)
		userAddressProtected.PATCH("/addresses/:id/default", h.User.UpdateDefaultUserAddress)
		userAddressProtected.PUT("/addresses/:id", h.User.UpdateUserAddress)
		userAddressProtected.DELETE("/addresses/:id", h.User.DeleteUserAddress)
	}

	{

		//admin auth routes
		adminAuth := g.Group("/api/v1/auth/admin/")
		adminAuth.POST("/login", h.Admin.LoginUser)
		adminAuth.POST("/refresh", h.Admin.RefreshToken)
		adminAuth.POST("/register", h.Admin.RegisterAdmin)

		adminProtectedRoute := g.Group("/api/v1/admin").Use(middleware.JWTMiddleware("admin", logger, cfg.JWTSecret))
		adminProtectedRoute.GET("/users", h.Admin.GetAllUsers)
		adminProtectedRoute.GET("/users/:id", h.Admin.GetUserByID)
		adminProtectedRoute.PUT("/users", h.Admin.UpdateUserStatus)
		adminProtectedRoute.DELETE("/users/:id", h.Admin.DeleteUser)
	}
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
		productProtectedRoute.PUT("/:id", h.Product.UpdateProduct)
		productProtectedRoute.PATCH("/:id", h.Product.UpdateProductStatus)
		productProtectedRoute.DELETE("/:id", h.Product.DeleteProduct)
	}

	{
		// Product variant routes
		productVariantRoute := g.Group("/api/v1/product-variants")
		productVariantOpenRoute := productVariantRoute.Group("/").Use(middleware.AuthContextMiddleware(logger, cfg.JWTSecret))
		productVariantOpenRoute.GET("/:id", h.Product.GetProductVariantByID)
		// productVariantOpenRoute.GET("/", h.Product.GetProductVariants)

		productVariantProtectedRoute := productVariantRoute.Group("/").Use(middleware.JWTMiddleware("admin", logger, cfg.JWTSecret))
		productVariantProtectedRoute.POST("/", h.Product.CreateProductVariant)
		productVariantProtectedRoute.PATCH("/:id", h.Product.UpdateVariantStatus)
		productVariantProtectedRoute.PUT("/:id", h.Product.UpdateProductVariant)
		productVariantProtectedRoute.DELETE("/:id", h.Product.DeleteProductVariant)

	}

	{
		// Variant attribute routes
		variantAttributeRoute := g.Group("/api/v1/attributes")
		variantAttributeProtectedRoute := variantAttributeRoute.Group("/").Use(middleware.JWTMiddleware("admin", logger, cfg.JWTSecret))
		variantAttributeProtectedRoute.POST("/", h.Product.CreateAttribute)
		// variantAttributeProtectedRoute.PUT("/", h.Product.UpdateVariantAttribute)
		variantAttributeProtectedRoute.DELETE("/:id", h.Product.DeleteAttribute)
		variantAttributeProtectedRoute.DELETE("/values", h.Product.DeleteAttributeValues)

		variantAttributeOpenRoute := variantAttributeRoute.Group("/").Use(middleware.AuthContextMiddleware(logger, cfg.JWTSecret))
		variantAttributeOpenRoute.GET("/", h.Product.GetAttributes)
		variantAttributeOpenRoute.GET("/:id", h.Product.GetAttributeByID)
	}

	{
		// cart routes
		cartRoute := g.Group("/api/v1/cart")
		cartProtectedRoute := cartRoute.Group("/").Use(middleware.JWTMiddleware("user", logger, cfg.JWTSecret))
		cartProtectedRoute.POST("/", h.Cart.AddToCart)
		cartProtectedRoute.GET("/", h.Cart.GetCart)
		cartProtectedRoute.PATCH("/", h.Cart.UpdateCartItemQuantity)
		cartProtectedRoute.DELETE("/items/:id", h.Cart.RemoveCartItem)
		cartProtectedRoute.DELETE("/", h.Cart.EmptyCart)
	}

	{
		// wishlist routes
		wishlistRoute := g.Group("/api/v1/wishlist")
		wishlistProtectedRoute := wishlistRoute.Group("/").Use(middleware.JWTMiddleware("user", logger, cfg.JWTSecret))
		wishlistProtectedRoute.POST("/", h.Wishlist.AddToWishlist)
		wishlistProtectedRoute.GET("/", h.Wishlist.GetWishlist)
		wishlistProtectedRoute.DELETE("/", h.Wishlist.RemoveFromWishlist)
		//wishlistProtectedRoute.DELETE("/", h.Wishlist.EmptyWishlist)
	}

	{
		// checkout routes
		checkoutRoute := g.Group("api/v1/checkout")
		checkoutProtectedRoute := checkoutRoute.Group("/").Use(middleware.JWTMiddleware("user", logger, cfg.JWTSecret))
		checkoutProtectedRoute.GET("/cart", h.Order.CheckoutCart)
		checkoutProtectedRoute.POST("/product-variant", h.Order.CheckoutProductVariant)
	}

}
