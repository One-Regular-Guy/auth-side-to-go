package settotp

import (
	"bytes"
	"context"
	"encoding/base64"
	"github.com/pquerna/otp/totp"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"image/png"
	"log"
	"os"
	"os/exec"
)

type Service interface {
	Collection() *mongo.Collection
	StepCaProvisioner() string
	StepCaProvisionerPassword() string
	GenTotpAndCert(uid string) (string, string, error)
}
type service struct {
	collection                *mongo.Collection
	stepCaProvisioner         string
	stepCaProvisionerPassword string
}

func (s *service) Collection() *mongo.Collection {
	return s.collection
}
func (s *service) StepCaProvisioner() string {
	return s.stepCaProvisioner
}
func (s *service) StepCaProvisionerPassword() string {
	return s.stepCaProvisionerPassword
}
func NewService(mongoCollection *mongo.Collection, stepCaProvisioner, stepCaProvisionerPassword string) Service {
	return &service{
		collection:                mongoCollection,
		stepCaProvisioner:         stepCaProvisioner,
		stepCaProvisionerPassword: stepCaProvisionerPassword,
	}
}

func (s *service) GenTotpAndCert(uid string) (string, string, error) {

	totpSecret, err := totp.Generate(
		totp.GenerateOpts{
			Issuer:      "AuthSideToGo",
			AccountName: "Maquiavelico",
		})
	if err != nil {
		return "", "", err
	}

	filter := bson.M{"_id": uid}
	update := bson.M{"$set": bson.M{"totp_secret": totpSecret}}
	opts := options.Update().SetUpsert(true)

	_, err = s.Collection().UpdateOne(context.Background(), filter, update, opts)
	if err != nil {
		return "", "", err
	}

	certDir := uid + "-cert.crt"
	certKey := uid + "-cert.key"
	err = os.Remove(certDir)
	if err != nil {
		log.Printf("Error Removing Previus Cert: %v", err)
	}
	err = os.Remove(certKey)
	if err != nil {
		log.Printf("Error Removing Previus Key: %v", err)
	}
	cmd := exec.Command("step", "ca", "certificate",
		uid,
		certDir,
		certKey,
		"--provisioner", s.StepCaProvisioner(),
		"--password-file", s.StepCaProvisionerPassword(),
		"-f",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Error when requesting cert: %v\nOutput: %s", err, output)
		return "", "", err
	}

	// Leitura dos arquivos gerados
	cert, err := os.ReadFile(certDir)
	if err != nil {
		log.Printf("Error when reading cert file: %v", err)
		return "", "", err
	}
	key, err := os.ReadFile(certKey)
	if err != nil {
		log.Printf("Error when Reading Key file: %v", err)
		return "", "", err
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
		log.Printf("Error when generating image: %v", err)
		return "", "", err
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		log.Printf("Error when encoding PNG: %v", err)
		return "", "", err
	}
	imgBase64 := base64.StdEncoding.EncodeToString(buf.Bytes())
	return imgBase64, ovpn, nil
}

var ServiceInstance Service
