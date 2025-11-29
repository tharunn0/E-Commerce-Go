package main

import (
	"context"
	"fmt"

	"github.com/tharunn0/E-Commerce-Go/internal/config"
	"github.com/tharunn0/E-Commerce-Go/internal/infrastructure/database"
	"github.com/tharunn0/E-Commerce-Go/internal/infrastructure/repository"
	"github.com/tharunn0/E-Commerce-Go/internal/presentation/handler"
	"github.com/tharunn0/E-Commerce-Go/internal/presentation/middleware"
	"github.com/tharunn0/E-Commerce-Go/internal/presentation/routes"
	"github.com/tharunn0/E-Commerce-Go/internal/service"
	"github.com/tharunn0/E-Commerce-Go/pkg/logger"
	"github.com/tharunn0/E-Commerce-Go/pkg/mailer"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	Razorpay "github.com/razorpay/razorpay-go"
	"go.uber.org/zap"
)

func main() {

	log := logger.InitLogger()
	ctx := context.Background()

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Failed to load .env")
	}
	cfg := config.LoadConfig()
	pg := cfg.Postgres
	smtp := cfg.SMTP
	app := cfg.App
	google := cfg.Google
	redis := cfg.Redis
	razorpay := cfg.Razorpay

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", pg.User, pg.Password, pg.Host, pg.Port, pg.DB)
	pgdb := database.InitDB(dsn, log)
	redisdb := database.InitRedis(ctx, redis, log)

	mailer, err := mailer.NewGoMailer(smtp.Port, smtp.Host, smtp.User, smtp.Pass, smtp.User)
	if err != nil {
		log.Fatal("Failed to initialize mailer", zap.String("function", "main"), zap.Error(err))
	}

	oauth := config.NewOAuthConfig(google.ClientID, google.ClientSecret, google.RedirectURL)
	razorpayClient := Razorpay.NewClient(razorpay.KeyID, razorpay.KeySecret)

	userRepo := repository.NewUserRepository(pgdb)
	authRepo := repository.NewAuthRepository(pgdb, redisdb)
	adminRepo := repository.NewAdminRepository(pgdb)
	categoryRepo := repository.NewCategoryRepository(pgdb)
	productRepo := repository.NewProductRepository(pgdb)
	cartRepo := repository.NewCartRepository(pgdb)
	wishlistRepo := repository.NewWishlistRepository(pgdb)
	orderRepo := repository.NewOrderRepository(pgdb)
	paymentRepo := repository.NewPaymentRepository(pgdb)

	userServ := service.NewUserService(userRepo, authRepo, log, mailer)
	adminServ := service.NewAdminService(adminRepo, log, authRepo)
	authServ := service.NewAuthService(authRepo, mailer, log, &cfg.Security)
	catergoryServ := service.NewCategoryService(categoryRepo, log)
	productServ := service.NewProductService(productRepo, log)
	cartServ := service.NewCartService(cartRepo, productRepo, log)
	wishlistServ := service.NewWishlistService(wishlistRepo, log)
	orderServ := service.NewOrderService(userRepo, productRepo, cartRepo, orderRepo, paymentRepo, razorpayClient, log)

	userHandler := handler.NewUserHandler(userServ, log, authServ, oauth)
	adminHandler := handler.NewAdminHandler(adminServ, log, authServ)
	categoryHandler := handler.NewCategoryHandler(catergoryServ, log)
	productHandler := handler.NewProductHandler(productServ, log)
	cartHandler := handler.NewCartHandler(cartServ, log)
	wishlistHandler := handler.NewWishlistHandler(wishlistServ, log)
	orderHandler := handler.NewOrderHandler(orderServ, log)
	paymentHandler := handler.NewPaymentHandler(&cfg.Razorpay, log)

	r := gin.New()
	r.Use(gin.Recovery(), middleware.RequestLogger(log))

	handler := routes.NewHandler(userHandler, adminHandler, categoryHandler, productHandler, cartHandler, wishlistHandler, orderHandler, paymentHandler)

	routes.RegisterRoutes(r, log, handler, &cfg.Security)

	log.Info(`Server starting at port : ` + app.Port)
	if er := r.Run(app.Host + ":" + app.Port); er != nil {
		log.Fatal("Server failed to start", zap.Error(er))
	}
}
