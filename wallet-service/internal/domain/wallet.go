package domain

import "errors"

// Common errors in the wallet domain.
var (
	ErrUserNotFound        = errors.New("user not found")
	ErrInsufficientBalance = errors.New("insufficient balance")
	ErrInvalidAmount       = errors.New("amount must be positive")
	ErrTransferToSelf      = errors.New("cannot transfer to self")
)

// Wallet represents the user's wallet entity.
type Wallet struct {
	ID      uint
	UserID  string
	Balance int64
}

// WalletRepository defines the interface for wallet data persistence.
type WalletRepository interface {
	// GetByUserID retrieves a wallet by its UserID.
	GetByUserID(userID string) (*Wallet, error)

	// Create persists a new wallet.
	Create(wallet *Wallet) error

	// UpdateBalance adds the amount to the user's balance safely using locking.
	UpdateBalance(userID string, amount int64) error

	// Transfer atomically moves funds from one user to another.
	Transfer(fromUser, toUser string, amount int64) error
}
