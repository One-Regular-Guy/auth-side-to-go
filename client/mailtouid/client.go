package main

import (
	"context"
	"github.com/One-Regular-Guy/auth-side-to-go/proto/mailtouid"
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
	mail := os.Getenv("MAIL")
	if mail == "" {
		log.Fatal("MAIL env variable is not set")
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
	client := mailtouid.NewMailToUidServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	req := &mailtouid.MailToUidRequest{Mail: mail}
	resp, err := client.MailToUid(ctx, req)
	if err != nil {
		log.Fatalf("Error calling MailToUid: %v", err)
	}

	log.Printf("Response: %v", resp)
}
