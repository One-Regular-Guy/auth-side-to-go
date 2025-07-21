package settotp

import (
	"go.mongodb.org/mongo-driver/mongo"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
)

type Service interface {
	Collection() *mongo.Collection
	StepCaUrl() string
	StepCaFingerprint() string
	StepCaProvisioner() string
	StepCaProvisionerPassword() string
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

func (s *service) GenTotpAndCert(uid string) (string, error) {

	certDir := uid + "/cert.crt"
	certKey := uid + "/cert.key"
	cmd := exec.Command("step", "ca", "certificate",
		"vpn.example.com", // Nome comum do certificado
		certDir,           // Arquivo de saída do certificado
		certKey,           // Arquivo de saída da chave privada
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
	cert, err := os.ReadFile("cert.crt")
	if err != nil {
		log.Fatalf("Erro ao ler arquivo de certificado: %v", err)
	}
	key, err := os.ReadFile("cert.key")
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
	err := os.WriteFile("client.ovpn", []byte(ovpn), 0600)
	if err != nil {
		log.Fatalf("Erro ao escrever arquivo .ovpn: %v", err)
	}
	return strings.TrimSpace(string(output)), nil
}

var ServiceInstance Service
