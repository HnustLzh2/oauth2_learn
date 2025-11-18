package ldap

import (
	"fmt"
	ldap "github.com/go-ldap/ldap/v3"
	"log"
	"net"
	"net/url"
	"oauth2/config"
	"strings"
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

func formatURL(ldapURL string) (string, error) {
	var protocol, hostport string
	_, err := url.Parse(ldapURL)
	if err != nil {
		return "", fmt.Errorf("parse Ldap HOST ERR: %s", err.Error())
	}
	if strings.Contains(ldapURL, "://") {
		splitLdapURL := strings.Split(ldapURL, "://")
		protocol, hostport = splitLdapURL[0], splitLdapURL[1]
		if !((protocol == "ldap") || (protocol == "ldaps")) {
			return "", fmt.Errorf("parse Ldap HOST ERR: protocol %s is not valid", protocol)
		}
	} else {
		hostport = ldapURL
		protocol = "ldap"
	}
	if strings.Contains(hostport, ":") {
		_, port, err := net.SplitHostPort(hostport)
		if err != nil {
			return "", fmt.Errorf("illegal ldap url, error: %v", err)
		}
		if port == "636" {
			protocol = "ldaps"
		}
	} else {
		switch protocol {
		case "ldap":
			hostport = hostport + ":389"
		case "ldaps":
			hostport = hostport + ":636"
		}
	}
	fLdapURL := protocol + "://" + hostport
	return fLdapURL, nil
}

func (s *Session) Open() error {
	ldapURL, err := formatURL(s.ldapCfg.URL)
	if err != nil {
		return err
	}
	log.Println(ldapURL)
	l, err := ldap.DialURL(ldapURL)
	if err != nil {
		return err
	}
	s.ldapConn = l
	return nil
}
