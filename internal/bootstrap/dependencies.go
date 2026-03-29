package bootstrap

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/razorpay/razorpay-go"
	"github.com/redis/go-redis/v9"
	"github.com/tharunn0/E-Commerce-Go/internal/config"
	"github.com/tharunn0/E-Commerce-Go/internal/infrastructure/database"
	"github.com/tharunn0/E-Commerce-Go/pkg/logger"
	"github.com/tharunn0/E-Commerce-Go/pkg/mailer"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

type Dependencies struct {
	RepoDeps
	ServiceDeps
	HandlerDeps
	ServerDeps
}

type RepoDeps struct {
	DB    *pgxpool.Pool
	Redis *redis.Client
}

type ServiceDeps struct {
	Logger   *zap.Logger
	Razorpay *razorpay.Client
	Mailer   *mailer.MailSender
	Cfg      *config.AppConfig
}

type HandlerDeps struct {
	Logger      *zap.Logger
	OAuthConfig *oauth2.Config
	Razorpay    *razorpay.Client
	Cfg         *config.AppConfig
	DB          *pgxpool.Pool
	Redis       *redis.Client
}

type ServerDeps struct {
	Logger *zap.Logger
	Cfg    *config.AppConfig
}

func InitDependencies() *Dependencies {

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

	return &Dependencies{
		RepoDeps: RepoDeps{
			DB:    pgdb,
			Redis: redisdb,
		},
		ServiceDeps: ServiceDeps{
			Logger:   log,
			Razorpay: razorpayClient,
			Mailer:   mailer,
			Cfg:      cfg,
		},
		HandlerDeps: HandlerDeps{
			Logger:      log,
			OAuthConfig: oauth,
			Razorpay:    razorpayClient,
			Cfg:         cfg,
			DB:          pgdb,
			Redis:       redisdb,
		},
		ServerDeps: ServerDeps{
			Logger: log,
			Cfg:    cfg,
		},
	}
}
