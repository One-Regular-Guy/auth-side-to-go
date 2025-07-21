package main

import (
	"context"
	"github.com/go-ldap/ldap/v3"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"log"
	"net"
	"os"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/One-Regular-Guy/auth-side-to-go/proto/mailtouid"
	"github.com/One-Regular-Guy/auth-side-to-go/proto/settotp"
)

type settotpServer struct {
	settotp.UnimplementedSetTotpServiceServer
}
type mailtouidServer struct {
	mailtouid.UnimplementedMailToUidServiceServer
}

func (s *mailtouidServer) MailToUid(ctx context.Context, req *mailtouid.MailToUidRequest) (*mailtouid.MailToUidResponse, error) {
	return mailtouid.RetrieveUidFromMail(req.Mail)
}
func (s *settotpServer) SetTotp(ctx context.Context, req *settotp.SetTotpRequest) (*settotp.SetTotpResponse, error) {
	return settotp.RetrieveTotpCodeFromMail(req.Uid)
}

func main() {
	// Load environment variables
	err := godotenv.Load()
	// Ldap Environment Variables
	baseDn := os.Getenv("LDAP_BASE_DN")
	bindDn := os.Getenv("LDAP_BIND_DN")
	password := os.Getenv("LDAP_BIND_PASSWORD")
	ldapUrl := os.Getenv("LDAP_URL")
	// Step & Mongo Environment Variables
	mongoUrl := os.Getenv("MONGO_URL")
	mongoUser := os.Getenv("MONGO_USER")
	mongoPassword := os.Getenv("MONGO_PASSWORD")
	mongoDb := os.Getenv("MONGO_DB")
	mongoTotpCollection := os.Getenv("MONGO_TOTP_COLLECTION")
	stepCaUrl := os.Getenv("STEP_CA_URL")
	stepCaFingerprint := os.Getenv("STEP_CA_FINGERPRINT")
	stepCaProvisioner := os.Getenv("STEP_CA_PROVISIONER")
	stepCaProvisionerPassword := os.Getenv("STEP_CA_PROVISIONER_PASSWORD")

	if err != nil {
		log.Fatalf("Erro ao carregar o arquivo .env: %v", err)
	}

	// Connect to LDAP server
	ldapConn, err := ldap.Dial("tcp", ldapUrl)
	if err != nil {
		log.Fatal(err)
	}
	defer func(ldapConn *ldap.Conn) {
		err := ldapConn.Close()
		if err != nil {

		}
	}(ldapConn)
	// Mongo connection
	// Conexão com o MongoDB
	mongoClient, err := mongo.Connect(context.Background(), options.Client().ApplyURI(mongoUrl).SetAuth(options.Credential{
		Username: mongoUser,
		Password: mongoPassword,
	}))
	if err != nil {
		log.Fatalf("Erro ao conectar ao MongoDB: %v", err)
	}
	defer func() {
		if err = mongoClient.Disconnect(context.Background()); err != nil {
			log.Fatalf("Erro ao desconectar do MongoDB: %v", err)
		}
	}()
	// Cria banco de dados e coleção
	db := mongoClient.Database(mongoDb)
	err = db.CreateCollection(context.Background(), mongoTotpCollection)
	if err != nil && !mongo.IsDuplicateKeyError(err) {
		log.Fatal(err)
	}
	collection := db.Collection(mongoTotpCollection)

	// Set up Necessary Business Logic Variables
	mailtouid.ServiceInstance = mailtouid.NewService(ldapConn, baseDn, bindDn, password)
	settotp.ServiceInstance = settotp.NewService(collection, stepCaUrl, stepCaFingerprint, stepCaProvisioner, stepCaProvisionerPassword)
	// Set up gRPC server with TLS
	certFile := "cert.pem"
	keyFile := "key.pem"
	creds, err := credentials.NewServerTLSFromFile(certFile, keyFile)
	if err != nil {
		log.Fatalf("Falha ao carregar certificados: %v", err)
	}

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Falha ao escutar: %v", err)
	}

	// TLS to gRPC server
	grpcServer := grpc.NewServer(grpc.Creds(creds))

	// Registra os serviços gRPC
	settotp.RegisterSetTotpServiceServer(grpcServer, &settotpServer{})
	mailtouid.RegisterMailToUidServiceServer(grpcServer, &mailtouidServer{})

	log.Println("Servidor gRPC rodando - :50051")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Falha ao servir: %v", err)
	}
}
