package ldap

import (
	ldap "github.com/go-ldap/ldap/v3"
	"oauth2/config"
)

type Session struct {
	ldapCfg  config.LDAP
	ldapConn *ldap.Conn
}

func NewSession(ldapCfg config.LDAP) *Session {
	return &Session{
		ldapCfg: ldapCfg,
	}
}

func fromURL(ldapURL string) (string, error) {

}
