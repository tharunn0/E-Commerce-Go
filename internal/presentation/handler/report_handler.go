package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tharunn0/E-Commerce-Go/internal/domain/report"
	"github.com/tharunn0/E-Commerce-Go/internal/service"
	"go.uber.org/zap"
)

type ReportHandler struct {
	serv *service.ReportService
	log  *zap.Logger
}

func NewReportHandler(serv *service.ReportService, log *zap.Logger) *ReportHandler {
	return &ReportHandler{
		serv: serv,
		log:  log,
	}
}

func (h *ReportHandler) GetSalesReport(c *gin.Context) {

	ctx := c.Request.Context()

	var req report.SalesReportRequest

	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "INVALID_REQUEST",
			"message": "invalid request",
		})
		return
	}

	reportData, apierr := h.serv.GetSalesReport(ctx, req)
	if apierr != nil {

		c.JSON(apierr.Status, gin.H{
			"error":   apierr.Code,
			"message": apierr.Message,
		})
		return

	}

	c.JSON(http.StatusOK, gin.H{
		"status":       "success",
		"sales_report": reportData,
	})

}

func (h *ReportHandler) GetTopSelling(c *gin.Context) {

	ctx := c.Request.Context()

	var req report.TopSellingRequest

	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "INVALID_REQUEST",
			"message": "invalid request",
		})
		return
	}

	resp, apierr := h.serv.GetTopSelling(ctx, req)
	if apierr != nil {

		c.JSON(apierr.Status, gin.H{
			"error":   apierr.Code,
			"message": apierr.Message,
		})
		return

	}

	c.JSON(http.StatusOK, gin.H{
		"status":      "success",
		"top_selling": resp,
	})

}

func (h *ReportHandler) GetRevenueAnalytics(c *gin.Context) {

	ctx := c.Request.Context()

	var req report.RevenueAnalyticsRequest

	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "INVALID_REQUEST",
			"message": "invalid request",
		})
		return
	}

	reportData, apierr := h.serv.GetRevenueAnalytics(ctx, req)
	if apierr != nil {

		c.JSON(apierr.Status, gin.H{
			"error":   apierr.Code,
			"message": apierr.Message,
		})
		return

	}

	c.JSON(http.StatusOK, gin.H{
		"status":            "success",
		"revenue_analytics": reportData,
	})

}
