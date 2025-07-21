package settotp

import (
	"bytes"
	"context"
	"encoding/base64"
	"github.com/pquerna/otp/totp"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"image/png"
	"log"
	"os"
	"os/exec"
)

type Service interface {
	Collection() *mongo.Collection
	StepCaUrl() string
	StepCaFingerprint() string
	StepCaProvisioner() string
	StepCaProvisionerPassword() string
	GenTotpAndCert(uid string) (string, string, error)
}
type service struct {
	collection                *mongo.Collection
	stepCaUrl                 string
	stepCaFingerprint         string
	stepCaProvisioner         string
	stepCaProvisionerPassword string
}

func (s *service) Collection() *mongo.Collection {
	return s.collection
}
func (s *service) StepCaUrl() string {
	return s.stepCaUrl
}
func (s *service) StepCaFingerprint() string {
	return s.stepCaFingerprint
}
func (s *service) StepCaProvisioner() string {
	return s.stepCaProvisioner
}
func (s *service) StepCaProvisionerPassword() string {
	return s.stepCaProvisionerPassword
}
func NewService(mongoCollection *mongo.Collection, stepCaUrl, stepCaFingerprint, stepCaProvisioner, stepCaProvisionerPassword string) Service {
	return &service{
		collection:                mongoCollection,
		stepCaUrl:                 stepCaUrl,
		stepCaFingerprint:         stepCaFingerprint,
		stepCaProvisioner:         stepCaProvisioner,
		stepCaProvisionerPassword: stepCaProvisionerPassword,
	}
}

func (s *service) GenTotpAndCert(uid string) (string, string, error) {

	totpSecret, err := totp.Generate(
		totp.GenerateOpts{
			Issuer:      "AuthSideToGo",
			AccountName: "VPN Coorporativa",
		})
	if err != nil {
		return "", "", err
	}

	_, err = s.Collection().InsertOne(context.Background(), bson.M{"_id": uid, "totp_secret": totpSecret})
	if err != nil {
		return "", "", err
	} // Exemplo de inserção de TOTP

	certDir := uid + "/cert.crt"
	certKey := uid + "/cert.key"
	cmd := exec.Command("step", "ca", "certificate",
		uid,     // Nome comum do certificado
		certDir, // Arquivo de saída do certificado
		certKey, // Arquivo de saída da chave privada
		"--ca-url", s.StepCaUrl(),
		"--fingerprint", s.StepCaFingerprint(),
		"--provisioner", s.StepCaProvisioner(),
		"--password", s.StepCaProvisionerPassword(), // Opcional, se necessário
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Erro ao solicitar certificado: %v\nSaída: %s", err, output)
	}

	// Leitura dos arquivos gerados
	cert, err := os.ReadFile(certDir)
	if err != nil {
		log.Fatalf("Erro ao ler arquivo de certificado: %v", err)
	}
	key, err := os.ReadFile(certKey)
	if err != nil {
		log.Fatalf("Erro ao ler arquivo de chave: %v", err)
	}

	// Exemplo de como incluir no arquivo .ovpn
	ovpn := `
<cert>
` + string(cert) + `
</cert>
<key>
` + string(key) + `
</key>
`
	img, err := totpSecret.Image(200, 200)
	if err != nil {
		return "", "", err
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return "", "", err
	}
	imgBase64 := base64.StdEncoding.EncodeToString(buf.Bytes())
	return ovpn, imgBase64, nil
}

var ServiceInstance Service
