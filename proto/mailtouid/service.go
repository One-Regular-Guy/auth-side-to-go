package mailtouid

import (
	"fmt"
	"log"

	"github.com/go-ldap/ldap/v3"
)

type Service interface {
	Conn() *ldap.Conn
	BaseDn() string
	BindDn() string
	Password() string
	Uid(email string) (string, error)
}
type service struct {
	conn     *ldap.Conn
	baseDn   string
	bindDn   string
	password string
}

func (s *service) Conn() *ldap.Conn {
	return s.conn
}
func (s *service) BaseDn() string {
	return s.baseDn
}
func (s *service) BindDn() string {
	return s.bindDn
}
func (s *service) Password() string {
	return s.password
}
func NewService(conn *ldap.Conn, base, bind, password string) Service {
	return &service{
		conn:     conn,
		baseDn:   base,
		bindDn:   bind,
		password: password,
	}
}
func (s *service) Uid(email string) (string, error) {
	presetErr := fmt.Errorf("usuário não encontrado")
	err := s.Conn().Bind(s.BindDn(), s.Password())
	if err != nil {
		log.Fatal(err)
	}
	searchRequest := ldap.NewSearchRequest(
		s.BaseDn(), // Base DN
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf("(mail=%s)", email), // Filtro
		[]string{"uid"},                 // Atributos a retornar
		nil,
	)

	sr, err := s.Conn().Search(searchRequest)
	if err != nil {
		log.Fatal(err)
	}

	if len(sr.Entries) == 0 {
		log.Print("Usuário não encontrado")
		return "", presetErr
	}
	uid := sr.Entries[0].GetAttributeValue("uid")
	return uid, nil
}

var ServiceInstance Service
