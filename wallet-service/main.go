package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	pb "pay-aja/proto/pb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Wallet represents the user's wallet entity stored in the database.
type Wallet struct {
	ID      uint   `gorm:"primaryKey"`
	UserID  string `gorm:"uniqueIndex;not null"`
	Balance int64  `gorm:"default:0"`
}

// server provides the implementation of the WalletServiceServer.
type server struct {
	pb.UnimplementedWalletServiceServer
	db *gorm.DB
}

// GetBalance retrieves the balance for a specific user.
// If the user does not exist, a new wallet is initialized with a balance of 0.
func (s *server) GetBalance(ctx context.Context, req *pb.GetBalanceRequest) (*pb.GetBalanceResponse, error) {
	var wallet Wallet

	result := s.db.Where("user_id = ?", req.GetUserId()).First(&wallet)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			wallet = Wallet{
				UserID:  req.GetUserId(),
				Balance: 0,
			}
			if err := s.db.Create(&wallet).Error; err != nil {
				return nil, status.Errorf(codes.Internal, "failed to create wallet: %v", err)
			}
		} else {
			return nil, status.Errorf(codes.Internal, "database error: %v", result.Error)
		}
	}

	return &pb.GetBalanceResponse{
		UserId:   wallet.UserID,
		Balance:  wallet.Balance,
		Currency: "IDR",
	}, nil
}

// TopUp increases the balance of a specific user.
// It returns an error if the amount is invalid or the user is not found.
func (s *server) TopUp(ctx context.Context, req *pb.TopUpRequest) (*pb.TopUpResponse, error) {
	if req.GetAmount() <= 0 {
		return nil, status.Errorf(codes.InvalidArgument, "amount must be positive")
	}

	var wallet Wallet
	if err := s.db.Where("user_id = ?", req.GetUserId()).First(&wallet).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "user not found")
		}
		return nil, status.Errorf(codes.Internal, "database query error: %v", err)
	}

	wallet.Balance += req.GetAmount()
	if err := s.db.Save(&wallet).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update balance: %v", err)
	}

	return &pb.TopUpResponse{
		UserId:  wallet.UserID,
		Balance: wallet.Balance,
		Status:  "SUCCESS",
	}, nil
}

func main() {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Jakarta",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	if err := db.AutoMigrate(&Wallet{}); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen on port 50051: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterWalletServiceServer(s, &server{db: db})

	log.Printf("Wallet Service is running on :50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve gRPC: %v", err)
	}
}
