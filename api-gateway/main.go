package main

import (
	"context"
	"log"
	"net/http"
	"time"

	pb "pay-aja/proto/pb"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// 1. Connect ke Wallet Service (gRPC)
	// Pastikan wallet-service jalan di port 50051
	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	defer conn.Close()

	client := pb.NewWalletServiceClient(conn)
	r := gin.Default()

	// --- Endpoint: Get Balance ---
	r.GET("/balance/:user_id", func(c *gin.Context) {
		userID := c.Param("user_id")

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		res, err := client.GetBalance(ctx, &pb.GetBalanceRequest{UserId: userID})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"user_id":  res.UserId,
			"balance":  res.Balance,
			"currency": res.Currency,
		})
	})

	// --- Endpoint: Top Up ---
	r.POST("/topup", func(c *gin.Context) {
		var req struct {
			UserID string `json:"user_id"`
			Amount int64  `json:"amount"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		res, err := client.TopUp(ctx, &pb.TopUpRequest{
			UserId: req.UserID,
			Amount: req.Amount,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":  res.Status,
			"balance": res.Balance,
		})
	})

	// --- Endpoint: Transfer ---
	r.POST("/transfer", func(c *gin.Context) {
		var req struct {
			FromUser string `json:"from_user"`
			ToUser   string `json:"to_user"`
			Amount   int64  `json:"amount"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		res, err := client.Transfer(ctx, &pb.TransferRequest{
			FromUserId: req.FromUser,
			ToUserId:   req.ToUser,
			Amount:     req.Amount,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":         res.Status,
			"transaction_id": res.TransactionId,
		})
	})

	log.Println("🔥 API Gateway running on :8080")
	r.Run(":8080")
}
