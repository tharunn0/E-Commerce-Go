package service

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/tharunn0/E-Commerce-Go/internal/apperror"
	"github.com/tharunn0/E-Commerce-Go/internal/domain"
	"go.uber.org/zap"
)

type ReportService struct {
	repo domain.ReportRepository
	log  *zap.Logger
}

func NewReportService(repo domain.ReportRepository, log *zap.Logger) *ReportService {
	return &ReportService{
		repo: repo,
		log:  log,
	}
}

func (s *ReportService) GetSalesReport(ctx context.Context, req domain.SalesReportRequest) (*domain.SalesReportResponse, *apperror.APIError) {

	log.Println("From: ", req.From)
	log.Println("To: ", req.To)

	// check if from is after 2020 and to is after from
	if err := req.Validate(time.Now()); err != nil {
		return nil, &apperror.APIError{
			Status:  http.StatusBadRequest,
			Code:    "INVALID_REQUEST",
			Message: err.Error(),
		}
	}

	reportResp, err := s.repo.GetSalesReport(ctx, &req)
	if err != nil {
		return nil, &apperror.APIError{
			Status:  http.StatusInternalServerError,
			Code:    "INTERNAL_ERROR",
			Message: "internal error",
		}
	}

	reportResp.SalesReportRequest = req

	return &reportResp, nil
}

func (s *ReportService) GetTopSelling(ctx context.Context, req domain.TopSellingRequest) (*domain.TopSellingResponse, *apperror.APIError) {

	//validate the request
	now := time.Now()
	if err := req.Validate(now); err != nil {
		s.log.Error("error validating top selling request", zap.Error(err))
		return nil, &apperror.APIError{
			Status:  http.StatusBadRequest,
			Code:    "INVALID_REQUEST",
			Message: err.Error(),
		}
	}

	// fetch top items
	var resp domain.TopSellingResponse
	var items []domain.TopStatItem
	var err error

	switch req.Type {
	case "product":
		items, err = s.repo.GetTopSellingProducts(ctx, &req)
	case "category":
		items, err = s.repo.GetTopSellingCategories(ctx, &req)
	case "brand":
		items, err = s.repo.GetTopSellingBrands(ctx, &req)
	}

	if err != nil {
		s.log.Error("error fetching top items", zap.Error(err))
		return nil, &apperror.APIError{
			Status:  http.StatusInternalServerError,
			Code:    "INTERNAL_ERROR",
			Message: "internal error",
		}
	}

	resp = domain.TopSellingResponse{
		Type:  req.Type,
		From:  req.From,
		To:    req.To,
		Items: items,
	}

	s.log.Info("top items fetched successfully")
	return &resp, nil
}

func (s *ReportService) GetRevenueAnalytics(ctx context.Context, req domain.RevenueAnalyticsRequest) (*domain.RevenueAnalyticsResponse, *apperror.APIError) {

	err := req.Validate(time.Now())
	if err != nil {
		s.log.Error("[service.GetRevenueAnalytics] error validating revenue analytics request", zap.Error(err))
		return nil, apperror.New(http.StatusBadRequest, "INVALID_REQUEST", err.Error())
	}

	repoResp, err := s.repo.GetRevenueAnalytics(ctx, &req)
	if err != nil {
		s.log.Error("[service.GetRevenueAnalytics] error fetching revenue analytics", zap.Error(err))
		return nil, apperror.New(http.StatusInternalServerError, "INTERNAL_ERROR", "internal error")
	}

	repoResp.RevenueAnalyticsRequest = req

	return repoResp, nil

}
