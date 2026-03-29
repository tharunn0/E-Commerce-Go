package bootstrap

import (
	"github.com/razorpay/razorpay-go"
	"github.com/tharunn0/E-Commerce-Go/internal/config"
	"github.com/tharunn0/E-Commerce-Go/internal/infrastructure/repository"
	"github.com/tharunn0/E-Commerce-Go/internal/presentation/handler"
	"github.com/tharunn0/E-Commerce-Go/internal/service"
	"github.com/tharunn0/E-Commerce-Go/pkg/mailer"
	"go.uber.org/zap"
)

type ExternalDeps struct {
	Logger   *zap.Logger
	Razorpay *razorpay.Client
	Mailer   *mailer.MailSender
	Cfg      *config.AppConfig
}

type Repositories struct {
	UserRepo     *repository.UserRepository
	AuthRepo     *repository.AuthRepository
	AdminRepo    *repository.AdminRepository
	CategoryRepo *repository.CategoryRepository
	ProductRepo  *repository.ProductRepository
	CartRepo     *repository.CartRepository
	WishlistRepo *repository.WishlistRepository
	OrderRepo    *repository.OrderRepository
	PaymentRepo  *repository.PaymentRepository
	OfferRepo    *repository.OfferRepository
	CouponRepo   *repository.CouponRepository
	ReportRepo   *repository.ReportRepository
	ReviewRepo   *repository.ReviewRepository
}

func NewRepositories(deps RepoDeps) *Repositories {
	return &Repositories{
		UserRepo:     repository.NewUserRepository(deps.DB),
		AuthRepo:     repository.NewAuthRepository(deps.DB, deps.Redis),
		AdminRepo:    repository.NewAdminRepository(deps.DB),
		CategoryRepo: repository.NewCategoryRepository(deps.DB),
		ProductRepo:  repository.NewProductRepository(deps.DB),
		CartRepo:     repository.NewCartRepository(deps.DB),
		WishlistRepo: repository.NewWishlistRepository(deps.DB),
		OrderRepo:    repository.NewOrderRepository(deps.DB),
		PaymentRepo:  repository.NewPaymentRepository(deps.DB),
		OfferRepo:    repository.NewOfferRepository(deps.DB),
		CouponRepo:   repository.NewCouponRepository(deps.DB),
		ReportRepo:   repository.NewReportRepository(deps.DB),
		ReviewRepo:   repository.NewReviewRepository(deps.DB),
	}
}

type Services struct {
	User     *service.UserService
	Admin    *service.AdminService
	Auth     *service.AuthService
	Category *service.CategoryService
	Product  *service.ProductService
	Cart     *service.CartService
	Wishlist *service.WishlistService
	Order    *service.OrderService
	Payment  *service.PaymentService
	Offer    *service.OfferService
	Coupon   *service.CouponService
	Report   *service.ReportService
	Review   *service.ReviewService
}

func NewServices(r *Repositories, deps ServiceDeps) *Services {
	return &Services{
		User:     service.NewUserService(r.UserRepo, r.AuthRepo, deps.Logger, deps.Mailer),
		Admin:    service.NewAdminService(r.AdminRepo, deps.Logger, r.AuthRepo),
		Auth:     service.NewAuthService(r.AuthRepo, deps.Mailer, deps.Logger, &deps.Cfg.Security),
		Category: service.NewCategoryService(r.CategoryRepo, deps.Logger),
		Product:  service.NewProductService(r.ProductRepo, r.OfferRepo, deps.Logger),
		Cart:     service.NewCartService(r.CartRepo, r.ProductRepo, r.OfferRepo, deps.Logger),
		Wishlist: service.NewWishlistService(r.WishlistRepo, deps.Logger),
		Order: service.NewOrderService(
			r.UserRepo, r.ProductRepo, r.CartRepo, r.OrderRepo,
			r.PaymentRepo, r.OfferRepo, r.CouponRepo,
			deps.Razorpay, deps.Cfg.Order, deps.Logger,
		),
		Payment: service.NewPaymentService(
			r.OrderRepo, r.PaymentRepo, r.UserRepo, r.CartRepo,
			deps.Razorpay, deps.Logger,
		),
		Offer:  service.NewOfferService(r.OfferRepo, deps.Logger),
		Coupon: service.NewCouponService(r.CouponRepo, deps.Logger, r.UserRepo, r.CartRepo, r.OfferRepo),
		Report: service.NewReportService(r.ReportRepo, deps.Logger),
		Review: service.NewReviewService(r.ReviewRepo, deps.Logger),
	}
}

type Handlers struct {
	User     *handler.UserHandler
	Admin    *handler.AdminHandler
	Category *handler.CategoryHandler
	Product  *handler.ProductHandler
	Cart     *handler.CartHandler
	Wishlist *handler.WishlistHandler
	Order    *handler.OrderHandler
	Payment  *handler.PaymentHandler
	Offer    *handler.OfferHandler
	Coupon   *handler.CouponHandler
	Report   *handler.ReportHandler
	Review   *handler.ReviewHandler
	Health   *handler.HealthHandler
}

func NewHandlers(s *Services, deps HandlerDeps) *Handlers {
	return &Handlers{
		User:     handler.NewUserHandler(s.User, s.Auth, deps.Logger, deps.OAuthConfig),
		Admin:    handler.NewAdminHandler(s.Admin, deps.Logger, s.Auth),
		Category: handler.NewCategoryHandler(s.Category, deps.Logger),
		Product:  handler.NewProductHandler(s.Product, deps.Logger),
		Cart:     handler.NewCartHandler(s.Cart, deps.Logger),
		Wishlist: handler.NewWishlistHandler(s.Wishlist, deps.Logger),
		Order:    handler.NewOrderHandler(s.Order, deps.Logger),
		Payment:  handler.NewPaymentHandler(s.Order, s.Payment, deps.Cfg.Razorpay, deps.Logger, deps.Razorpay),
		Offer:    handler.NewOfferHandler(s.Offer, deps.Logger),
		Coupon:   handler.NewCouponHandler(s.Coupon, deps.Logger),
		Report:   handler.NewReportHandler(s.Report, deps.Logger),
		Review:   handler.NewReviewHandler(s.Review, deps.Logger),
		Health:   handler.NewHealthHandler(deps.DB, deps.Redis, deps.Logger),
	}
}
