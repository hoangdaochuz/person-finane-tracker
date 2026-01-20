package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/dev/personal-finance-tracker/backend/internal/domain"
	"github.com/dev/personal-finance-tracker/backend/internal/service"
)

// WebhookHandler handles webhook requests from iOS app
type WebhookHandler struct {
	service service.TransactionService
}

// NewWebhookHandler creates a new webhook handler
func NewWebhookHandler(service service.TransactionService) *WebhookHandler {
	return &WebhookHandler{service: service}
}

// CreateTransaction handles single transaction creation via webhook
func (h *WebhookHandler) CreateTransaction(c *gin.Context) {
	var req domain.CreateTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	transaction, err := h.service.CreateTransaction(&req)
	if err != nil {
		// Check if it's a validation error
		var validationErr *domain.ValidationError
		if errors.As(err, &validationErr) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": validationErr.Message,
				"field": validationErr.Field,
			})
			return
		}

		// All other errors are internal server errors
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}

	c.JSON(http.StatusCreated, transaction)
}

// CreateBatchTransaction handles batch transaction creation via webhook
func (h *WebhookHandler) CreateBatchTransaction(c *gin.Context) {
	var req domain.BatchTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	transactions, err := h.service.CreateBatchTransaction(&req)
	if err != nil {
		// Check if it's a validation error
		var validationErr *domain.ValidationError
		if errors.As(err, &validationErr) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": validationErr.Message,
				"field": validationErr.Field,
			})
			return
		}

		// All other errors are internal server errors
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"created":      len(transactions),
		"transactions": transactions,
	})
}
