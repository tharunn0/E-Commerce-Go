package main

import (
	"context"
	"fmt"
	"os"

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

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Failed to load .env")
	}

	db_str := os.Getenv("PGDB_URL")
	pgdb := database.InitDB(db_str, log)
	pgdb.Query(context.Background(), "")

	mailer := mailer.NewGoMailer(587, os.Getenv("EMAIL_HOST"), os.Getenv("EMAIL_USERNAME"), os.Getenv("EMAIL_PASSWORD"), os.Getenv("EMAIL"))

	authRepo := repository.NewAuthRepository(pgdb)
	userRepo := repository.NewUserRepository(pgdb)
	adminRepo := repository.NewAdminRepository(pgdb)
	categoryRepo := repository.NewCategoryRepository(pgdb)
	productRepo := repository.NewProductRepository(pgdb)

	userServ := service.NewUserService(userRepo, authRepo, log, mailer)
	adminServ := service.NewAdminService(adminRepo, log)
	authServ := service.NewAuthService(authRepo, mailer, log)
	catergoryServ := service.NewCategoryService(categoryRepo, log)
	productServ := service.NewProductService(productRepo, log)

	userHandler := handler.NewUserHandler(userServ, log, authServ)
	adminHandler := handler.NewAdminHandler(adminServ, log)
	categoryHandler := handler.NewCategoryHandler(catergoryServ, log)
	productHandler := handler.NewProductHandler(productServ, log)

	r := gin.New()
	r.Use(gin.Recovery(), gin.Logger(), middleware.RequestLogger(log))

	routes.RegisterRoutes(r, log, userHandler, adminHandler, categoryHandler, productHandler)

	port := 8080
	log.Info(`Server starting at port : ` + fmt.Sprintf("%d", port))
	if er := r.Run(fmt.Sprintf("127.0.0.1:%d", port)); er != nil {
		log.Fatal("Server failed to start", zap.Error(er))
	}
}
