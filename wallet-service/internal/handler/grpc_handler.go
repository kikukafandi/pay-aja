package handler

import (
	"context"
	"fmt"
	pb "pay-aja/proto/pb"
	"pay-aja/wallet-service/internal/domain"
	"pay-aja/wallet-service/internal/usecase"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// WalletHandler implements the gRPC WalletServiceServer interface.
type WalletHandler struct {
	pb.UnimplementedWalletServiceServer
	usecase usecase.WalletUsecase
}

// NewWalletHandler creates a new gRPC handler for wallet service.
func NewWalletHandler(u usecase.WalletUsecase) *WalletHandler {
	return &WalletHandler{usecase: u}
}

func (h *WalletHandler) GetBalance(ctx context.Context, req *pb.GetBalanceRequest) (*pb.GetBalanceResponse, error) {
	wallet, err := h.usecase.GetBalance(req.GetUserId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get balance: %v", err)
	}

	return &pb.GetBalanceResponse{
		UserId:   wallet.UserID,
		Balance:  wallet.Balance,
		Currency: "IDR",
	}, nil
}

func (h *WalletHandler) TopUp(ctx context.Context, req *pb.TopUpRequest) (*pb.TopUpResponse, error) {
	wallet, err := h.usecase.TopUp(req.GetUserId(), req.GetAmount())
	if err != nil {
		switch err {
		case domain.ErrUserNotFound:
			return nil, status.Error(codes.NotFound, err.Error())
		case domain.ErrInvalidAmount:
			return nil, status.Error(codes.InvalidArgument, err.Error())
		default:
			return nil, status.Errorf(codes.Internal, "topup failed: %v", err)
		}
	}

	return &pb.TopUpResponse{
		UserId:  wallet.UserID,
		Balance: wallet.Balance,
		Status:  "SUCCESS",
	}, nil
}

func (h *WalletHandler) Transfer(ctx context.Context, req *pb.TransferRequest) (*pb.TransferResponse, error) {
	err := h.usecase.Transfer(req.GetFromUserId(), req.GetToUserId(), req.GetAmount())
	if err != nil {
		switch err {
		case domain.ErrUserNotFound:
			return nil, status.Error(codes.NotFound, err.Error())
		case domain.ErrInsufficientBalance:
			return nil, status.Error(codes.FailedPrecondition, err.Error())
		case domain.ErrInvalidAmount, domain.ErrTransferToSelf:
			return nil, status.Error(codes.InvalidArgument, err.Error())
		default:
			return nil, status.Errorf(codes.Internal, "transfer failed: %v", err)
		}
	}

	return &pb.TransferResponse{
		Status:        "SUCCESS",
		TransactionId: fmt.Sprintf("%s-to-%s", req.GetFromUserId(), req.GetToUserId()),
	}, nil
}
