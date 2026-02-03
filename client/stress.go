package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	pb "pay-aja/proto/pb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// 1. Setup Koneksi ke gRPC Server
	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Gagal connect ke server: %v", err)
	}
	defer conn.Close()

	client := pb.NewWalletServiceClient(conn)

	// 2. Parameter Test
	userID := "user-stress-test"
	totalRequest := 1000        // Jumlah serangan
	amountPerReq := int64(1000) // Nominal per topup
	expectedTotal := int64(totalRequest) * amountPerReq

	// 3. Cek Saldo Awal
	ctx := context.Background()
	initResp, err := client.GetBalance(ctx, &pb.GetBalanceRequest{UserId: userID})
	if err != nil {
		log.Fatalf("Gagal ambil saldo awal (pastikan server nyala): %v", err)
	}
	initialBalance := initResp.Balance
	fmt.Printf("\nSaldo Awal %s: %d\n", userID, initialBalance)

	fmt.Printf("MEMULAI Test: %d Request x Rp %d...\n", totalRequest, amountPerReq)
	fmt.Println("   (Mohon tunggu sebentar...)")

	start := time.Now()
	var wg sync.WaitGroup
	wg.Add(totalRequest)

	// 4. Eksekusi 1000 Goroutines (Concurrent Request)
	for i := 0; i < totalRequest; i++ {
		go func() {
			defer wg.Done()

			// Beri timeout per request biar tidak hang kalau server sibuk
			subCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			_, err := client.TopUp(subCtx, &pb.TopUpRequest{
				UserId: userID,
				Amount: amountPerReq,
			})

			if err != nil {
				// Kalau error, cetak 'x' kecil biar tau ada yang gagal
				fmt.Print("x")
			}
		}()
	}

	wg.Wait() // Tunggu semua goroutine selesai
	duration := time.Since(start)

	// 5. Cek Hasil Akhir
	finalResp, err := client.GetBalance(ctx, &pb.GetBalanceRequest{UserId: userID})
	if err != nil {
		log.Fatalf("Gagal cek saldo akhir: %v", err)
	}

	actualTotalInc := finalResp.Balance - initialBalance

	// 6. Laporan
	fmt.Println("\n\n--- LAPORAN STRESS TEST ---")
	fmt.Printf("Waktu Eksekusi: %s\n", duration)
	fmt.Printf("Total Masuk: Rp %d\n", actualTotalInc)
	fmt.Printf("Ekspektasi : Rp %d\n", expectedTotal)

	if actualTotalInc == expectedTotal {
		fmt.Println("SISTEM AMAN! (Tidak ada Race Condition)")
	} else {
		fmt.Printf("BAHAYA! HILANG Rp %d (Terjadi Race Condition)\n", expectedTotal-actualTotalInc)
		fmt.Println("Analisis: Database gagal menangani request yang masuk bersamaan.")
	}
}
