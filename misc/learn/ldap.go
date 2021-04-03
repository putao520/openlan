package main

import (
	"fmt"
	"github.com/go-ldap/ldap"
	"os"
)

type LDAPConfig struct {
	Addr         string
	BindUserName string
	BindPassword string
	SearchDN     string
}

type LDAPService struct {
	Conn   *ldap.Conn
	Config LDAPConfig
}

func NewLDAPService(config LDAPConfig) (*LDAPService, error) {
	conn, err := ldap.Dial("tcp", config.Addr)
	if err != nil {
		return nil, err
	}

	// NOTE(chenjun): skip verify
	// err = conn.StartTLS(&tls.Config{InsecureSkipVerify: true})
	// if err != nil {
	//  return nil, err
	// }

	err = conn.Bind(config.BindUserName, config.BindPassword)
	if err != nil {
		return nil, err
	}

	return &LDAPService{Conn: conn, Config: config}, nil
}

func (l *LDAPService) Login(userName, password string) (bool, error) {
	searchRequest := ldap.NewSearchRequest(
		l.Config.SearchDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf("(&(objectClass=inetOrgPerson)(mail=%s))", userName),
		[]string{"dn"},
		nil,
	)

	result, err := l.Conn.Search(searchRequest)
	if err != nil {
		return false, err
	}

	if len(result.Entries) != 1 {
		return false, fmt.Errorf("find user entries %d", len(result.Entries))
	}

	userDN := result.Entries[0].DN
	err = l.Conn.Bind(userDN, password)
	if err != nil {
		return false, err
	}

	err = l.Conn.Bind(l.Config.BindUserName, l.Config.BindPassword)
	if err != nil {
		return false, nil
	}

	return true, nil
}

func main() {
	cfg := LDAPConfig{}
	cfg.Addr = os.Getenv("LDAPServer")
	cfg.BindPassword = os.Getenv("LADPPassword")
	cfg.BindUserName = os.Getenv("LADPBaseDN")
	cfg.SearchDN = os.Getenv("LADPSearchDN")

	if ldap, err := NewLDAPService(cfg); err != nil {
		panic(err)
	} else {
		username := os.Getenv("username")
		password := os.Getenv("password")
		if ok, err := ldap.Login(username, password); !ok {
			panic(err)
		} else {
			fmt.Println("success")
		}
	}
	fmt.Println(cfg)
}
