package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/mantis-exchange/mantis-wallet/internal/service"
)

type Handler struct {
	wallet *service.WalletService
}

func New(wallet *service.WalletService) *Handler {
	return &Handler{wallet: wallet}
}

func (h *Handler) GetDepositAddress(c *gin.Context) {
	userID, err := uuid.Parse(c.Query("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}
	chain := c.DefaultQuery("chain", "ETH")

	addr, err := h.wallet.GetOrCreateDepositAddress(c.Request.Context(), userID, chain)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"address": addr.Address, "chain": addr.Chain})
}

type withdrawReq struct {
	UserID  string `json:"user_id" binding:"required"`
	Chain   string `json:"chain" binding:"required"`
	Asset   string `json:"asset" binding:"required"`
	Address string `json:"address" binding:"required"`
	Amount  string `json:"amount" binding:"required"`
}

func (h *Handler) RequestWithdrawal(c *gin.Context) {
	var req withdrawReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	w, err := h.wallet.RequestWithdrawal(c.Request.Context(), userID, req.Chain, req.Asset, req.Address, req.Amount, "0")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"withdrawal_id": w.ID, "status": w.Status})
}
