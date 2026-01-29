package main

import (
	"context"
	"log"
	"net"

	pb "pay-aja/proto/pb"

	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedWalletServiceServer
}

func (s *server) GetBalance(ctx context.Context, req *pb.GetBalanceRequest) (*pb.GetBalanceResponse, error) {
	return &pb.GetBalanceResponse{
		UserId:   req.GetUserId(),
		Balance:  15000000,
		Currency: "IDR",
	}, nil
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterWalletServiceServer(s, &server{})

	log.Printf("🚀 PayAja running on :50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
