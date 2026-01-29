package main

import (
	"context"
	"log"
	"time"

	pb "pay-aja/proto/pb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// 1. Hubungi Server di localhost:50051
	// Kita pakai "insecure" karena di local belum pakai SSL/TLS
	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	c := pb.NewWalletServiceClient(conn)

	// 2. Siapkan Context (Timeout 1 detik)
	// Agar kalau server mati, client tidak bengong selamanya
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// 3. Panggil Fungsi "GetBalance"
	// Skenario: User baru dengan ID "user-coba-1" (Harusnya dibuatkan wallet saldo 0)
	targetUser := "user-coba-1"
	log.Printf("Mencoba ambil saldo untuk: %s...", targetUser)

	r, err := c.GetBalance(ctx, &pb.GetBalanceRequest{UserId: targetUser})
	if err != nil {
		log.Fatalf("Error saat request: %v", err)
	}

	// 4. Tampilkan Hasil
	log.Printf("✅ SUKSES! User ID: %s | Saldo: %d %s", r.GetUserId(), r.GetBalance(), r.GetCurrency())
}
