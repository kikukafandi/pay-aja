package repository

import (
	"pay-aja/wallet-service/internal/domain"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// WalletModel represents the database schema for wallets.
type WalletModel struct {
	ID      uint   `gorm:"primaryKey"`
	UserID  string `gorm:"uniqueIndex;not null"`
	Balance int64  `gorm:"default:0"`
}

// TableName overrides the table name to 'wallets'.
func (WalletModel) TableName() string {
	return "wallets"
}

// walletRepository implements domain.WalletRepository using GORM.
type walletRepository struct {
	db *gorm.DB
}

// NewWalletRepository creates a new instance of walletRepository.
func NewWalletRepository(db *gorm.DB) domain.WalletRepository {
	db.AutoMigrate(&WalletModel{})
	return &walletRepository{db: db}
}

// GetByUserID retrieves a wallet by UserID. Returns nil if not found.
func (r *walletRepository) GetByUserID(userID string) (*domain.Wallet, error) {
	var model WalletModel
	err := r.db.Where("user_id = ?", userID).First(&model).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &domain.Wallet{
		ID:      model.ID,
		UserID:  model.UserID,
		Balance: model.Balance,
	}, nil
}

// Create inserts a new wallet record.
func (r *walletRepository) Create(w *domain.Wallet) error {
	model := WalletModel{
		UserID:  w.UserID,
		Balance: w.Balance,
	}
	return r.db.Create(&model).Error
}

// UpdateBalance updates the wallet balance within a transaction using row locking.
func (r *walletRepository) UpdateBalance(userID string, amount int64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var model WalletModel
		// Use optimistic locking to prevent race conditions.
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("user_id = ?", userID).
			First(&model).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return domain.ErrUserNotFound
			}
			return err
		}

		model.Balance += amount
		return tx.Save(&model).Error
	})
}

// Transfer performs an atomic fund transfer between two users with locking.
func (r *walletRepository) Transfer(fromUser, toUser string, amount int64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var fromWallet, toWallet WalletModel

		// Lock sender
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("user_id = ?", fromUser).First(&fromWallet).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return domain.ErrUserNotFound
			}
			return err
		}

		if fromWallet.Balance < amount {
			return domain.ErrInsufficientBalance
		}

		// Lock receiver
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("user_id = ?", toUser).First(&toWallet).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return domain.ErrUserNotFound
			}
			return err
		}

		fromWallet.Balance -= amount
		toWallet.Balance += amount

		if err := tx.Save(&fromWallet).Error; err != nil {
			return err
		}
		return tx.Save(&toWallet).Error
	})
}
