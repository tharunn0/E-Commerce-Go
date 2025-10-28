package main

import (
	"context"
	"fmt"

	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/tharunn0/E-Commerce-Go/internal/infrastructure/database"
	"github.com/tharunn0/E-Commerce-Go/internal/infrastructure/repository"
	"github.com/tharunn0/E-Commerce-Go/internal/presentation/handler"
	"github.com/tharunn0/E-Commerce-Go/internal/presentation/middleware"
	"github.com/tharunn0/E-Commerce-Go/internal/presentation/routes"
	"github.com/tharunn0/E-Commerce-Go/internal/service"
	"github.com/tharunn0/E-Commerce-Go/pkg/logger"
	"github.com/tharunn0/E-Commerce-Go/pkg/mailer"

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

	authrepo := repository.NewAuthRepository(pgdb)
	userRepo := repository.NewUserRepository(pgdb)

	userServ := service.NewUserService(userRepo, authrepo, log, mailer)
	authServ := service.NewAuthService(authrepo, mailer, log)

	userHandler := handler.NewUserHandler(userServ, log, authServ)

	r := gin.New()
	r.Use(gin.Recovery(), gin.Logger(), middleware.RequestLogger(log))

	routes.RegisterRoutes(r, log, userHandler)

	port := 8080
	log.Info(`Server starting at port : ` + fmt.Sprintf("%d", port))
	if er := r.Run(fmt.Sprintf("127.0.0.1:%d", port)); er != nil {
		log.Fatal("Server failed to start", zap.Error(er))
	}
}
