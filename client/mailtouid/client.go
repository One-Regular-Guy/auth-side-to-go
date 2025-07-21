package main

import (
	"context"
	"github.com/One-Regular-Guy/auth-side-to-go/proto/mailtouid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"log"
	"time"
)

func main() {
	// Carrega as credenciais TLS
	creds := credentials.NewClientTLSFromCert(nil, "")

	// Conecta ao servidor
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(creds))
	if err != nil {
		log.Fatalf("Falha ao conectar: %v", err)
	}
	defer conn.Close()

	// Cria o client
	client := mailtouid.NewMailToUidServiceClient(conn)

	// Exemplo de chamada RPC
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Substitua por um request válido
	req := &mailtouid.MailToUidRequest{Mail: "exemplo@exemplo.com"}
	resp, err := client.MailToUid(ctx, req)
	if err != nil {
		log.Fatalf("Erro ao chamar MailToUid: %v", err)
	}

	log.Printf("Resposta: %v", resp)
}
