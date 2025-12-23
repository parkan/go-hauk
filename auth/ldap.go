package auth

import (
	"fmt"
	"strings"

	"github.com/go-ldap/ldap/v3"
)

type LDAPAuth struct {
	uri        string
	baseDN     string
	bindDN     string
	bindPass   string
	userFilter string
	startTLS   bool
}

func NewLDAPAuth(uri, baseDN, bindDN, bindPass, userFilter string, startTLS bool) *LDAPAuth {
	return &LDAPAuth{
		uri:        uri,
		baseDN:     baseDN,
		bindDN:     bindDN,
		bindPass:   bindPass,
		userFilter: userFilter,
		startTLS:   startTLS,
	}
}

func (l *LDAPAuth) Authenticate(user, pass string) error {
	if pass == "" {
		return ErrAuthFailed
	}

	conn, err := ldap.DialURL(l.uri)
	if err != nil {
		return fmt.Errorf("ldap connect: %w", err)
	}
	defer conn.Close()

	if l.startTLS {
		if err := conn.StartTLS(nil); err != nil {
			return fmt.Errorf("ldap starttls: %w", err)
		}
	}

	if err := conn.Bind(l.bindDN, l.bindPass); err != nil {
		return fmt.Errorf("ldap admin bind: %w", err)
	}

	filter := strings.Replace(l.userFilter, "%s", ldap.EscapeFilter(user), 1)
	req := ldap.NewSearchRequest(
		l.baseDN,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		0, 0, false,
		filter,
		[]string{"dn"},
		nil,
	)

	res, err := conn.Search(req)
	if err != nil {
		return fmt.Errorf("ldap search: %w", err)
	}

	if len(res.Entries) == 0 {
		return ErrAuthFailed
	}
	if len(res.Entries) > 1 {
		return fmt.Errorf("ldap: ambiguous user filter matched %d users", len(res.Entries))
	}

	userDN := res.Entries[0].DN
	if err := conn.Bind(userDN, pass); err != nil {
		return ErrAuthFailed
	}

	return nil
}
