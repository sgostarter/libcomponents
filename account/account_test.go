package account

import (
	// nolint
	"crypto/md5"
	"testing"

	"github.com/sgostarter/i/l"
	"github.com/stretchr/testify/assert"
)

// nolint
func TestAccountToken(t *testing.T) {
	account := &accountImpl{
		logger:  l.NewConsoleLoggerWrapper(),
		storage: nil,
		cfg: &Config{
			PasswordHashIterCount: 1024,
		},
	}

	tokenKey := md5.Sum([]byte("abcd"))
	account.tokenKey = tokenKey[:]

	token, err := account.tokenNew(10, "user10")
	assert.Nil(t, err)

	uid, userName, err := account.tokenCheck(token)
	assert.Nil(t, err)
	assert.EqualValues(t, 10, uid)
	assert.EqualValues(t, "user10", userName)
}
