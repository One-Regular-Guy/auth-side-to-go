package mailtouid

import "fmt"

func RetrieveUidFromMail(mail string) (*MailToUidResponse, error) {
	if mail == "" {
		return nil, fmt.Errorf("fudeu")
	}
	uid, err := ServiceInstance.Uid(mail)
	if err != nil {
		return nil, err
	}
	return &MailToUidResponse{Uid: uid}, nil
}
