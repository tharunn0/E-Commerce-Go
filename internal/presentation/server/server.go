package server

import (
	"github.com/gin-gonic/gin"
	"github.com/tharunn0/E-Commerce-Go/internal/bootstrap"
	"github.com/tharunn0/E-Commerce-Go/internal/presentation/middleware"
	"go.uber.org/zap"
)

func StartServer(handlers *bootstrap.Handlers, deps bootstrap.ServerDeps) {

	handler := NewRouteHandler(handlers.User, handlers.Admin, handlers.Category, handlers.Product,
		handlers.Cart, handlers.Wishlist, handlers.Order, handlers.Payment, handlers.Offer, handlers.Coupon, handlers.Report)

	r := gin.New()
	r.Use(gin.Recovery(), middleware.RequestLogger(deps.Logger))

	RegisterRoutes(r, deps.Logger, handler, &deps.Cfg.Security)

	deps.Logger.Info(`Server starting at port : ` + deps.Cfg.App.Port)
	if er := r.Run(deps.Cfg.App.Host + ":" + deps.Cfg.App.Port); er != nil {
		deps.Logger.Fatal("Server failed to start", zap.Error(er))
	}
}
