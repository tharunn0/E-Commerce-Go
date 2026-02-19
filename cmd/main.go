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
	"github.com/razorpay/razorpay-go"
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

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", cfg.Postgres.User, cfg.Postgres.Password, cfg.Postgres.Host, cfg.Postgres.Port, cfg.Postgres.DB)
	pgdb := database.InitDB(dsn, log)
	redisdb := database.InitRedis(ctx, cfg.Redis, log)

	mailer, err := mailer.NewGoMailer(cfg.SMTP.Port, cfg.SMTP.Host, cfg.SMTP.User, cfg.SMTP.Pass, cfg.SMTP.User)
	if err != nil {
		log.Fatal("Failed to initialize mailer", zap.String("function", "main"), zap.Error(err))
	}

	oauth := config.NewOAuthConfig(cfg.Google.ClientID, cfg.Google.ClientSecret, cfg.Google.RedirectURL)
	razorpayClient := razorpay.NewClient(cfg.Razorpay.KeyID, cfg.Razorpay.KeySecret)

	userRepo := repository.NewUserRepository(pgdb)
	authRepo := repository.NewAuthRepository(pgdb, redisdb)
	adminRepo := repository.NewAdminRepository(pgdb)
	categoryRepo := repository.NewCategoryRepository(pgdb)
	productRepo := repository.NewProductRepository(pgdb)
	cartRepo := repository.NewCartRepository(pgdb)
	wishlistRepo := repository.NewWishlistRepository(pgdb)
	orderRepo := repository.NewOrderRepository(pgdb)
	paymentRepo := repository.NewPaymentRepository(pgdb)
	offerRepo := repository.NewOfferRepository(pgdb)
	couponRepo := repository.NewCouponRepository(pgdb)
	reportRepo := repository.NewReportRepository(pgdb)

	userServ := service.NewUserService(userRepo, authRepo, log, mailer)
	adminServ := service.NewAdminService(adminRepo, log, authRepo)
	authServ := service.NewAuthService(authRepo, mailer, log, &cfg.Security)
	catergoryServ := service.NewCategoryService(categoryRepo, log)
	productServ := service.NewProductService(productRepo, offerRepo, log)
	cartServ := service.NewCartService(cartRepo, productRepo, offerRepo, log)
	wishlistServ := service.NewWishlistService(wishlistRepo, log)
	orderServ := service.NewOrderService(userRepo, productRepo, cartRepo, orderRepo, paymentRepo, offerRepo, razorpayClient, log)
	paymentServ := service.NewPaymentService(orderRepo, paymentRepo, userRepo, razorpayClient, log)
	offerServ := service.NewOfferService(offerRepo, log)
	couponServ := service.NewCouponService(couponRepo, log, userRepo, cartRepo, offerRepo)
	reportServ := service.NewReportService(reportRepo, log)

	userHandler := handler.NewUserHandler(userServ, log, authServ, oauth)
	adminHandler := handler.NewAdminHandler(adminServ, log, authServ)
	categoryHandler := handler.NewCategoryHandler(catergoryServ, log)
	productHandler := handler.NewProductHandler(productServ, log)
	cartHandler := handler.NewCartHandler(cartServ, log)
	wishlistHandler := handler.NewWishlistHandler(wishlistServ, log)
	orderHandler := handler.NewOrderHandler(orderServ, log)
	paymentHandler := handler.NewPaymentHandler(orderServ, paymentServ, cfg.Razorpay, log, razorpayClient)
	offerHandler := handler.NewOfferHandler(offerServ, log)
	couponHandler := handler.NewCouponHandler(couponServ, log)
	reportHandler := handler.NewReportHandler(reportServ, log)

	r := gin.New()
	r.Use(gin.Recovery(), middleware.RequestLogger(log))

	handler := routes.NewHandler(userHandler, adminHandler, categoryHandler, productHandler,
		cartHandler, wishlistHandler, orderHandler, paymentHandler, offerHandler, couponHandler, reportHandler)

	routes.RegisterRoutes(r, log, handler, &cfg.Security)

	log.Info(`Server starting at port : ` + cfg.App.Port)
	if er := r.Run(cfg.App.Host + ":" + cfg.App.Port); er != nil {
		log.Fatal("Server failed to start", zap.Error(er))
	}

}
