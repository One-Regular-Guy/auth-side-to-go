package settotp

import "fmt"

func RetrieveTotpCodeFromMail(uid string) (*SetTotpResponse, error) {
	if uid == "" {
		return nil, fmt.Errorf("fudeu")
	}
	code := "code_for_" + uid
	return &SetTotpResponse{Code: code}, nil
}
