package usecase

import (
	"pay-aja/wallet-service/internal/domain"
)

// WalletUsecase defines the business logic for wallet operations.
type WalletUsecase interface {
	GetBalance(userID string) (*domain.Wallet, error)
	TopUp(userID string, amount int64) (*domain.Wallet, error)
	Transfer(fromUser, toUser string, amount int64) error
}

type walletUsecase struct {
	repo domain.WalletRepository
}

// NewWalletUsecase creates a new instance of walletUsecase.
func NewWalletUsecase(repo domain.WalletRepository) WalletUsecase {
	return &walletUsecase{repo: repo}
}

// GetBalance retrieves the balance. Creates a new wallet if the user does not exist.
func (u *walletUsecase) GetBalance(userID string) (*domain.Wallet, error) {
	w, err := u.repo.GetByUserID(userID)
	if err != nil {
		return nil, err
	}

	if w == nil {
		newWallet := &domain.Wallet{
			UserID:  userID,
			Balance: 0,
		}
		if err := u.repo.Create(newWallet); err != nil {
			return nil, err
		}
		return newWallet, nil
	}

	return w, nil
}

// TopUp adds funds to a user's wallet.
func (u *walletUsecase) TopUp(userID string, amount int64) (*domain.Wallet, error) {
	if amount <= 0 {
		return nil, domain.ErrInvalidAmount
	}

	if err := u.repo.UpdateBalance(userID, amount); err != nil {
		return nil, err
	}

	return u.repo.GetByUserID(userID)
}

// Transfer moves funds from one user to another.
func (u *walletUsecase) Transfer(fromUser, toUser string, amount int64) error {
	if amount <= 0 {
		return domain.ErrInvalidAmount
	}
	if fromUser == toUser {
		return domain.ErrTransferToSelf
	}

	return u.repo.Transfer(fromUser, toUser, amount)
}
