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
	"os/exec"

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
	actualHost := os.Getenv("HOST")
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
	stepCaProvisionerPasswordFile := os.Getenv("STEP_CA_PROVISIONER_PASSWORD")

	if err != nil {
		log.Println("Error loading environment variables")
	}

	// Bootsrap CA
	cmd := exec.Command("step", "ca", "bootstrap",
		"--ca-url", stepCaUrl,
		"--fingerprint", stepCaFingerprint,
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Error when connecting to CA to bootstrap: %v\nOutput: %s", err, output)
	}
	err = os.Remove("cert.pem")
	if err != nil {
		log.Printf("Erro when removing previus cert: %v", err)
	}
	err = os.Remove("key.pem")
	if err != nil {
		log.Printf("Erro when removing previus key: %v", err)
	}
	cmd = exec.Command("step", "ca", "certificate",
		actualHost,
		"cert.pem",
		"key.pem",
		"--provisioner", stepCaProvisioner,
		"--password-file", stepCaProvisionerPasswordFile,
		"-f",
	)
	output, err = cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Error when connecting to CA to gen certs: %v\nOutput: %s", err, output)
	}
	// Connect to LDAP server
	ldapConn, err := ldap.DialURL(ldapUrl)
	if err != nil {
		log.Fatal(err)
	}
	defer func(ldapConn *ldap.Conn) {
		err := ldapConn.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(ldapConn)
	// Connect to MongoDB
	mongoClient, err := mongo.Connect(context.Background(), options.Client().ApplyURI(mongoUrl).SetAuth(options.Credential{
		Username: mongoUser,
		Password: mongoPassword,
	}))
	if err != nil {
		log.Fatalf("Error when connection to MongoDB: %v", err)
	}
	defer func() {
		if err = mongoClient.Disconnect(context.Background()); err != nil {
			log.Fatalf("Error when disconnecting MongoDB: %v", err)
		}
	}()
	// Create Database and Collection
	db := mongoClient.Database(mongoDb)
	collection := db.Collection(mongoTotpCollection)

	// Set up Necessary Business Logic Variables
	mailtouid.ServiceInstance = mailtouid.NewService(ldapConn, baseDn, bindDn, password)
	settotp.ServiceInstance = settotp.NewService(collection, stepCaProvisioner, stepCaProvisionerPasswordFile)

	certFile := "cert.pem"
	keyFile := "key.pem"
	creds, err := credentials.NewServerTLSFromFile(certFile, keyFile)
	if err != nil {
		log.Fatalf("Failed to gen creds: %v", err)
	}

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer(grpc.Creds(creds))

	settotp.RegisterSetTotpServiceServer(grpcServer, &settotpServer{})
	mailtouid.RegisterMailToUidServiceServer(grpcServer, &mailtouidServer{})

	log.Println("Service gRPC running - :50051")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Falha ao servir: %v", err)
	}
}
