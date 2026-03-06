package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/tharunn0/E-Commerce-Go/internal/service"
	"go.uber.org/zap"
)

type ReviewHandler struct {
	service *service.ReviewService
	log     *zap.Logger
}

func NewReviewHandler(service *service.ReviewService, log *zap.Logger) *ReviewHandler {
	return &ReviewHandler{service: service, log: log}
}

func (h *ReviewHandler) CreateReview(ctx *gin.Context) {

}

func (h *ReviewHandler) DeleteReview(ctx *gin.Context) {

}
