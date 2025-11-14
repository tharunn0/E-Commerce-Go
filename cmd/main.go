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

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", pg.User, pg.Password, pg.Host, pg.Port, pg.DB)
	pgdb := database.InitDB(dsn, log)
	redisdb := database.InitRedis(ctx, redis, log)

	mailer, err := mailer.NewGoMailer(smtp.Port, smtp.Host, smtp.User, smtp.Pass, smtp.User)
	if err != nil {
		log.Fatal("Failed to initialize mailer", zap.String("function", "main"), zap.Error(err))
	}

	oauth := config.NewOAuthConfig(google.ClientID, google.ClientSecret, google.RedirectURL)

	userRepo := repository.NewUserRepository(pgdb)
	authRepo := repository.NewAuthRepository(pgdb, redisdb)
	adminRepo := repository.NewAdminRepository(pgdb)
	categoryRepo := repository.NewCategoryRepository(pgdb)
	productRepo := repository.NewProductRepository(pgdb)

	userServ := service.NewUserService(userRepo, authRepo, log, mailer)
	adminServ := service.NewAdminService(adminRepo, log)
	authServ := service.NewAuthService(authRepo, mailer, log, &cfg.Security)
	catergoryServ := service.NewCategoryService(categoryRepo, log)
	productServ := service.NewProductService(productRepo, log)

	userHandler := handler.NewUserHandler(userServ, log, authServ, oauth)
	adminHandler := handler.NewAdminHandler(adminServ, log)
	categoryHandler := handler.NewCategoryHandler(catergoryServ, log)
	productHandler := handler.NewProductHandler(productServ, log)

	r := gin.New()
	r.Use(gin.Recovery(), middleware.RequestLogger(log))

	handler := routes.NewHandler(userHandler, adminHandler, categoryHandler, productHandler)

	routes.RegisterRoutes(r, log, handler, &cfg.Security)

	log.Info(`Server starting at port : ` + app.Port)
	if er := r.Run(app.Host + ":" + app.Port); er != nil {
		log.Fatal("Server failed to start", zap.Error(er))
	}
}
