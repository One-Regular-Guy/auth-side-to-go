package settotp

import "fmt"

func RetrieveTotpCodeFromMail(uid string) (*SetTotpResponse, error) {
	if uid == "" {
		return nil, fmt.Errorf("Cannot set TOTP code: UID is empty")
	}
	code, cert, err := ServiceInstance.GenTotpAndCert(uid)
	if err != nil {
		return nil, fmt.Errorf("Error generating TOTP code and certificate: %v", err)
	}
	return &SetTotpResponse{Code: code, Cert: cert}, nil
}
