package fixtures

import (
	"errors"
	"strings"

	kallax "github.com/src-d/go-kallax"
)

type User struct {
	kallax.Model `table:"users"`
	Username     string
	Email        string
	Password     Password
	Websites     []string
	Emails       []*Email
	Settings     *Settings
}

func newUser(username, email string) (*User, error) {
	if strings.Contains(email, "@spam.org") {
		return nil, errors.New("kallax: is spam!")
	}
	return &User{Username: username, Email: email}, nil
}

type Email struct {
	kallax.Model `table:"emails"`
	Address      string
	Primary      bool
}

func newProfile(address string, primary bool) *Email {
	return &Email{Address: address, Primary: primary}
}

type Password string

// Kids, don't do this at home
func (p *Password) Set(pwd string) {
	*p = Password("such cypher" + pwd + "much secure")
}

type Settings struct {
	NotificationsActive bool
	NotifyByEmail       bool
}
