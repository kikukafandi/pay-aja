package main

import (
	"context"
	"testing"

	pb "pay-aja/proto/pb"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB inisialisasi DB bersih setiap dipanggil
func setupTestDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	db.AutoMigrate(&Wallet{})
	return db
}

func TestTopUp(t *testing.T) {
	// Definisi Tabel Kasus (Scenarios)
	tests := []struct {
		name        string
		initBalance int64
		amount      int64
		wantBalance int64
		wantError   bool
	}{
		{
			name:        "Positive_Normal",
			initBalance: 0,
			amount:      50000,
			wantBalance: 50000,
			wantError:   false,
		},
		{
			name:        "Positive_ExistingBalance",
			initBalance: 10000,
			amount:      20000,
			wantBalance: 30000,
			wantError:   false,
		},
		{
			name:        "Negative_Amount_ShouldFail",
			initBalance: 50000,
			amount:      -1000,
			wantBalance: 50000, // Saldo tidak boleh berubah
			wantError:   true,
		},
	}

	// Eksekusi Loop
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db := setupTestDB()
			s := &server{db: db}
			userID := "user-topup"

			// Seed Data Awal
			db.Create(&Wallet{UserID: userID, Balance: tc.initBalance})

			// Action
			resp, err := s.TopUp(context.Background(), &pb.TopUpRequest{
				UserId: userID,
				Amount: tc.amount,
			})

			// Assertions
			if tc.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.wantBalance, resp.Balance)

				// Double Check ke DB
				var w Wallet
				db.Where("user_id = ?", userID).First(&w)
				assert.Equal(t, tc.wantBalance, w.Balance)
			}
		})
	}
}

func TestTransfer(t *testing.T) {
	// Tabel Skenario Transfer yang Komplit
	tests := []struct {
		name         string
		balanceA     int64
		balanceB     int64
		transferAmnt int64
		wantBalanceA int64
		wantBalanceB int64
		wantError    bool
		wantStatus   string
	}{
		{
			name:         "Success_Normal",
			balanceA:     100000,
			balanceB:     0,
			transferAmnt: 20000,
			wantBalanceA: 80000,
			wantBalanceB: 20000,
			wantError:    false,
			wantStatus:   "SUCCESS",
		},
		{
			name:         "Fail_InsufficientBalance",
			balanceA:     10000,
			balanceB:     0,
			transferAmnt: 50000,
			wantBalanceA: 10000, // Tidak berkurang
			wantBalanceB: 0,     // Tidak bertambah
			wantError:    true,
			wantStatus:   "",
		},
		{
			name:         "Fail_NegativeAmount",
			balanceA:     50000,
			balanceB:     0,
			transferAmnt: -5000,
			wantBalanceA: 50000,
			wantBalanceB: 0,
			wantError:    true,
			wantStatus:   "",
		},
		{
			name:         "Fail_TransferToSelf",
			balanceA:     50000,
			balanceB:     0, // Diabaikan karena transfer ke diri sendiri
			transferAmnt: 10000,
			wantBalanceA: 50000,
			wantBalanceB: 0,
			wantError:    true,
			wantStatus:   "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db := setupTestDB()
			s := &server{db: db}

			userA := "user-A"
			userB := "user-B"

			// Seed Data
			db.Create(&Wallet{UserID: userA, Balance: tc.balanceA})
			db.Create(&Wallet{UserID: userB, Balance: tc.balanceB})

			// Handle Self Transfer Case (A ke A)
			toUser := userB
			if tc.name == "Fail_TransferToSelf" {
				toUser = userA
			}

			// Action
			resp, err := s.Transfer(context.Background(), &pb.TransferRequest{
				FromUserId: userA,
				ToUserId:   toUser,
				Amount:     tc.transferAmnt,
			})

			// Assertions
			if tc.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.wantStatus, resp.Status)
			}

			// Verifikasi Saldo Akhir di DB
			var wa, wb Wallet
			db.Where("user_id = ?", userA).First(&wa)
			// Hanya cek B jika bukan transfer ke diri sendiri
			if tc.name != "Fail_TransferToSelf" {
				db.Where("user_id = ?", userB).First(&wb)
				assert.Equal(t, tc.wantBalanceB, wb.Balance, "Saldo Penerima Salah")
			}
			assert.Equal(t, tc.wantBalanceA, wa.Balance, "Saldo Pengirim Salah")
		})
	}
}
