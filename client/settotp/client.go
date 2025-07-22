package main

import (
	"context"
	"github.com/One-Regular-Guy/auth-side-to-go/proto/settotp"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"log"
	"os"
	"time"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading dot env: %v", err)
	}
	remoteAddr := os.Getenv("REMOTE_ADDR")
	if remoteAddr == "" {
		log.Fatal("REMOTE_ADDR env variable is not set")
	}
	uid := os.Getenv("UID")
	if uid == "" {
		log.Fatal("UID env variable is not set")
	}
	// Load TLS credentials
	creds := credentials.NewClientTLSFromCert(nil, "")

	// Create a connection to the server
	conn, err := grpc.NewClient(remoteAddr+":50051", grpc.WithTransportCredentials(creds))
	if err != nil {
		log.Fatalf("Create Connection Failed: %v", err)
	}
	defer func(conn *grpc.ClientConn) {
		err := conn.Close()
		if err != nil {
			log.Fatalf("Close Connection Failed: %v", err)
		}
	}(conn)

	client := settotp.NewSetTotpServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req := &settotp.SetTotpRequest{Uid: uid}
	resp, err := client.SetTotp(ctx, req)
	if err != nil {
		log.Fatalf("Error calling SetTotp: %v", err)
	}

	log.Printf("Response: %v", resp)
}
