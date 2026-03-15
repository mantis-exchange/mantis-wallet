package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/mantis-exchange/mantis-wallet/internal/chain"
	"github.com/mantis-exchange/mantis-wallet/internal/config"
	"github.com/mantis-exchange/mantis-wallet/internal/handler"
	"github.com/mantis-exchange/mantis-wallet/internal/model"
	"github.com/mantis-exchange/mantis-wallet/internal/service"
)

func main() {
	cfg := config.Load()

	pool, err := pgxpool.New(context.Background(), cfg.DBURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	repo := model.NewWalletRepo(pool)
	eth := chain.NewEthereumClient(cfg.ETHNode)
	walletService := service.NewWalletService(repo, eth)

	// Start deposit scanner
	go eth.ScanDeposits()

	// Process approved withdrawals periodically
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			if err := walletService.ProcessPendingWithdrawals(context.Background()); err != nil {
				log.Printf("withdrawal processing error: %v", err)
			}
		}
	}()

	h := handler.New(walletService)
	r := gin.Default()
	api := r.Group("/api/v1/wallet")
	{
		api.GET("/deposit-address", h.GetDepositAddress)
		api.POST("/withdraw", h.RequestWithdrawal)
	}

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	log.Printf("mantis-wallet starting on :%s", cfg.Port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}
