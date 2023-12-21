package account

import (
	"crypto/md5" // nolint: gosec
	"fmt"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sgostarter/libeasygo/cuserror"
	"github.com/stretchr/testify/assert"
)

func TestJWTExpiredAt(t *testing.T) {
	tokenKey := md5.Sum([]byte("x")) // nolint: gosec
	tokenKeyR := tokenKey[:]

	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			//ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Second)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(tokenKeyR)
	assert.Nil(t, err)
	assert.True(t, token != "")

	time.Sleep(time.Second * 2)

	var claims2 Claims

	tokenD, err := jwt.ParseWithClaims(token, &claims2, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, cuserror.NewWithErrorMsg(fmt.Sprintf("Unexpected signing method: %v", token.Header["alg"]))
		}

		return tokenKeyR, nil
	}, /*jwt.WithExpirationRequired(),*/ jwt.WithIssuedAt())

	assert.Nil(t, err)
	assert.True(t, tokenD.Valid)
}
