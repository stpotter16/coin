package authorization

import (
	"errors"

	"github.com/stpotter16/coin/internal/types"
)

type Authorizer struct {
	passphrase string
}

func New(getenv func(string) string) (Authorizer, error) {
	passphrase := getenv("BIODATA_PASSPHRASE")
	if passphrase == "" {
		return Authorizer{}, errors.New("could not locate passphrase environment variable")
	}

	a := Authorizer{
		passphrase: passphrase,
	}

	return a, nil
}

func (a Authorizer) Authorize(loginRequest types.LoginRequest) bool {
	return loginRequest.Passphrase == a.passphrase
}

func (a Authorizer) AuthorizeApi(headerVal string) bool {
	return headerVal == a.passphrase
}
