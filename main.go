package main

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"log"
	"net"

	// Importe os pacotes gerados
	"seu/modulo/proto/mailtouid"
	"seu/modulo/proto/settotp"
)

type settotpServer struct {
	settotp.UnimplementedSetTotpServiceServer
}
type mailtouidServer struct {
	mailtouid.UnimplementedMailToUidServiceServer
}

func main() {
	// Caminhos para o certificado e chave privada
	certFile := "cert/server.crt"
	keyFile := "cert/server.key"

	// Carrega as credenciais TLS
	creds, err := credentials.NewServerTLSFromFile(certFile, keyFile)
	if err != nil {
		log.Fatalf("Falha ao carregar certificados: %v", err)
	}

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Falha ao escutar: %v", err)
	}

	// Cria o servidor gRPC com TLS
	grpcServer := grpc.NewServer(grpc.Creds(creds))

	// Registra os serviços
	settotp.RegisterSetTotpServiceServer(grpcServer, &settotpServer{})
	mailtouid.RegisterMailToUidServiceServer(grpcServer, &mailtouidServer{})

	log.Println("Servidor gRPC com TLS rodando na porta 50051")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Falha ao servir: %v", err)
	}
}
