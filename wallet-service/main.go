package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"time"

	pb "pay-aja/proto/pb"
	"pay-aja/wallet-service/internal/handler"
	"pay-aja/wallet-service/internal/repository"
	"pay-aja/wallet-service/internal/usecase"

	"google.golang.org/grpc"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// 1. Initialize Database Connection
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

	// Configure Connection Pooling
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("failed to get sql.DB: %v", err)
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(50)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// 2. Setup Dependency Injection
	walletRepo := repository.NewWalletRepository(db)
	walletUC := usecase.NewWalletUsecase(walletRepo)
	walletHandler := handler.NewWalletHandler(walletUC)

	// 3. Start gRPC Server
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterWalletServiceServer(s, walletHandler)

	log.Printf("Wallet Service (DDD) running on :50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
