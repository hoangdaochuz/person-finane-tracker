package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/dev/personal-finance-tracker/backend/internal/domain"
	"github.com/dev/personal-finance-tracker/backend/internal/service"
)

// AnalyticsHandler handles analytics and dashboard requests
type AnalyticsHandler struct {
	service service.TransactionService
}

// NewAnalyticsHandler creates a new analytics handler
func NewAnalyticsHandler(service service.TransactionService) *AnalyticsHandler {
	return &AnalyticsHandler{service: service}
}

// GetSummary returns the financial summary (total in/out, balance)
func (h *AnalyticsHandler) GetSummary(c *gin.Context) {
	summary, err := h.service.GetSummary()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, summary)
}

// GetTrends returns trends over time (daily, weekly, monthly)
func (h *AnalyticsHandler) GetTrends(c *gin.Context) {
	period := c.DefaultQuery("period", "daily")

	if period != "daily" && period != "weekly" && period != "monthly" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "period must be one of: daily, weekly, monthly",
		})
		return
	}

	trends, err := h.service.GetTrends(period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, trends)
}

// GetBreakdownBySource returns breakdown of expenses by source (bank/wallet)
func (h *AnalyticsHandler) GetBreakdownBySource(c *gin.Context) {
	breakdown, err := h.service.GetBreakdownBySource()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, breakdown)
}

// GetBreakdownByCategory returns breakdown of expenses by category
func (h *AnalyticsHandler) GetBreakdownByCategory(c *gin.Context) {
	breakdown, err := h.service.GetBreakdownByCategory()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, breakdown)
}

// ListTransactions returns a paginated list of transactions
func (h *AnalyticsHandler) ListTransactions(c *gin.Context) {
	var params domain.ListTransactionsQueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	transactions, total, err := h.service.ListTransactions(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": transactions,
		"pagination": gin.H{
			"page":        params.Page,
			"page_size":   params.PageSize,
			"total":       total,
			"total_pages": (total + int64(params.PageSize) - 1) / int64(params.PageSize),
		},
	})
}

// GetTransactionByID returns a single transaction by ID
func (h *AnalyticsHandler) GetTransactionByID(c *gin.Context) {
	id := c.Param("id")
	var transactionID int64
	if _, err := fmt.Sscanf(id, "%d", &transactionID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid transaction ID",
		})
		return
	}

	transaction, err := h.service.GetTransactionByID(transactionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, transaction)
}
