package service

import (
	"context"
	"net/http"
	"time"

	"github.com/tharunn0/E-Commerce-Go/internal/apperror"
	"github.com/tharunn0/E-Commerce-Go/internal/domain/report"
	"go.uber.org/zap"
)

type ReportService struct {
	repo report.ReportRepository
	log  *zap.Logger
}

func NewReportService(repo report.ReportRepository, log *zap.Logger) *ReportService {
	return &ReportService{
		repo: repo,
		log:  log,
	}
}

func (s *ReportService) GetSalesReport(ctx context.Context, req report.SalesReportRequest) (*report.SalesReportResponse, *apperror.APIError) {

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
		s.log.Error("error fetching sales report", zap.Error(err))
		return nil, &apperror.APIError{
			Status:  http.StatusInternalServerError,
			Code:    "INTERNAL_ERROR",
			Message: "internal error",
		}
	}

	reportResp.SalesReportRequest = req

	return &reportResp, nil
}

func (s *ReportService) GetTopSelling(ctx context.Context, req report.TopSellingRequest) (*report.TopSellingResponse, *apperror.APIError) {

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
	var resp report.TopSellingResponse
	var items []report.TopStatItem
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

	resp = report.TopSellingResponse{
		Type:  req.Type,
		From:  req.From,
		To:    req.To,
		Items: items,
	}

	s.log.Info("top items fetched successfully")
	return &resp, nil
}

func (s *ReportService) GetRevenueAnalytics(ctx context.Context, req report.RevenueAnalyticsRequest) (*report.RevenueAnalyticsResponse, *apperror.APIError) {

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

func (s *ReportService) GetDashboard(ctx context.Context, req *report.DashboardRequest) (*report.DashboardResponse, *apperror.APIError) {

	now := time.Now()
	if err := req.Validate(now); err != nil {
		s.log.Error("[service.GetDashboard] error validating dashboard request", zap.Error(err))
		return nil, apperror.New(http.StatusBadRequest, "INVALID_REQUEST", err.Error())
	}

	resp, err := s.repo.GetDashboard(ctx, req)
	if err != nil {
		s.log.Error("[service.GetDashboard] error fetching dashboard", zap.Error(err))
		return nil, apperror.New(http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch dashboard")
	}

	// fetch summary

	reportFilter := report.SalesReportRequest{
		From: req.From,
		To:   req.To,
	}
	summary, err := s.repo.GetSalesReport(ctx, &reportFilter)
	if err != nil {
		s.log.Error("")
	}

	resp.Summary = &summary.SalesSummary

	// fetch top products. categories and brands

	reqFilter := &report.TopSellingRequest{
		From:  req.From,
		To:    req.To,
		Limit: 5,
	}

	products, err := s.repo.GetTopSellingProducts(ctx, reqFilter)
	if err != nil {
		s.log.Error("[service.GetDashboard] error fetching top products", zap.Error(err))
		return nil, apperror.New(http.StatusInternalServerError, "INTERNAL_ERROR", "internal error")
	}

	resp.TopProducts = products

	categories, err := s.repo.GetTopSellingCategories(ctx, reqFilter)
	if err != nil {
		s.log.Error("[service.GetDashboard] error fetching top categories", zap.Error(err))
		return nil, apperror.New(http.StatusInternalServerError, "INTERNAL_ERROR", "internal error")
	}

	resp.TopCategories = categories

	brands, err := s.repo.GetTopSellingBrands(ctx, reqFilter)
	if err != nil {
		s.log.Error("[service.GetDashboard] error fetching top brands", zap.Error(err))
		return nil, apperror.New(http.StatusInternalServerError, "INTERNAL_ERROR", "internal error")
	}

	resp.TopBrands = brands

	resp.From = req.From
	resp.To = req.To

	return resp, nil
}
