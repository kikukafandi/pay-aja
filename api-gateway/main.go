package main

import (
	"context"
	"log"
	"net/http"
	"time"

	pb "pay-aja/proto/pb"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	_ "pay-aja/api-gateway/docs"
)

// --- STRUCT UNTUK DOKUMENTASI SWAGGER ---

// TopUpInput mendefinisikan payload untuk topup
type TopUpInput struct {
	UserID string `json:"user_id" example:"sultan-1"`
	Amount int64  `json:"amount" example:"100000"`
}

// TransferInput mendefinisikan payload untuk transfer
type TransferInput struct {
	FromUser string `json:"from_user" example:"sultan-1"`
	ToUser   string `json:"to_user" example:"rakyat-1"`
	Amount   int64  `json:"amount" example:"50000"`
}

// BalanceResponse model response saldo
type BalanceResponse struct {
	UserID   string `json:"user_id" example:"sultan-1"`
	Balance  int64  `json:"balance" example:"1000000"`
	Currency string `json:"currency" example:"IDR"`
}

// StatusResponse model response sukses transaksi
type StatusResponse struct {
	Status        string `json:"status" example:"SUCCESS"`
	TransactionID string `json:"transaction_id,omitempty" example:"txn-123"`
	Balance       int64  `json:"balance,omitempty" example:"1000000"`
}

// ErrorResponse model response error
type ErrorResponse struct {
	Error string `json:"error" example:"user not found"`
}

// ----------------------------------------

type GatewayHandler struct {
	client pb.WalletServiceClient
}

// @title           PayAja Wallet API
// @version         1.0
// @description     API Gateway untuk layanan E-Wallet (gRPC Backend).
// @host            localhost:8080
// @BasePath        /
func main() {
	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	defer conn.Close()

	client := pb.NewWalletServiceClient(conn)
	h := &GatewayHandler{client: client}

	r := gin.Default()

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	r.GET("/balance/:user_id", h.GetBalance)
	r.POST("/topup", h.TopUp)
	r.POST("/transfer", h.Transfer)

	log.Println("🔥 API Gateway running on :8080")
	log.Println("📄 Swagger UI: http://localhost:8080/swagger/index.html")
	r.Run(":8080")
}

// @Summary      Cek Saldo
// @Description  Mengambil saldo user.
// @Tags         Wallet
// @Accept       json
// @Produce      json
// @Param        user_id   path      string  true  "User ID"
// @Success      200       {object}  BalanceResponse
// @Failure      500       {object}  ErrorResponse
// @Router       /balance/{user_id} [get]
func (h *GatewayHandler) GetBalance(c *gin.Context) {
	userID := c.Param("user_id")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	res, err := h.client.GetBalance(ctx, &pb.GetBalanceRequest{UserId: userID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, BalanceResponse{
		UserID:   res.UserId,
		Balance:  res.Balance,
		Currency: res.Currency,
	})
}

// @Summary      Top Up Saldo
// @Description  Menambah saldo user.
// @Tags         Wallet
// @Accept       json
// @Produce      json
// @Param        request   body      TopUpInput  true  "Data Top Up"
// @Success      200       {object}  StatusResponse
// @Failure      400       {object}  ErrorResponse
// @Failure      500       {object}  ErrorResponse
// @Router       /topup [post]
func (h *GatewayHandler) TopUp(c *gin.Context) {
	var req TopUpInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid input"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	res, err := h.client.TopUp(ctx, &pb.TopUpRequest{
		UserId: req.UserID,
		Amount: req.Amount,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, StatusResponse{
		Status:  res.Status,
		Balance: res.Balance,
	})
}

// @Summary      Transfer Uang
// @Description  Transfer saldo antar user.
// @Tags         Wallet
// @Accept       json
// @Produce      json
// @Param        request   body      TransferInput  true  "Data Transfer"
// @Success      200       {object}  StatusResponse
// @Failure      400       {object}  ErrorResponse
// @Failure      500       {object}  ErrorResponse
// @Router       /transfer [post]
func (h *GatewayHandler) Transfer(c *gin.Context) {
	var req TransferInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid input"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	res, err := h.client.Transfer(ctx, &pb.TransferRequest{
		FromUserId: req.FromUser,
		ToUserId:   req.ToUser,
		Amount:     req.Amount,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, StatusResponse{
		Status:        res.Status,
		TransactionID: res.TransactionId,
	})
}
