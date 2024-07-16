package account

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sgostarter/i/commerr"
	"github.com/sgostarter/libeasygo/cuserror"
)

type Claims struct {
	UID      uint64 `json:"uid"`
	UserName string `json:"userName"`
	jwt.RegisteredClaims
}

func (impl *accountImpl) tokenNew(uid uint64, userName string) (token string, err error) {
	claims := Claims{
		UID:      uid,
		UserName: userName,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token, err = jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(impl.tokenKey)

	return
}

func (impl *accountImpl) tokenCheck(tokenS string) (uid uint64, userName string, err error) {
	var claims Claims

	token, err := jwt.ParseWithClaims(tokenS, &claims, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, cuserror.NewWithErrorMsg(fmt.Sprintf("Unexpected signing method: %v", token.Header["alg"]))
		}

		return impl.tokenKey, nil
	})

	if err != nil {
		return
	}

	if !token.Valid {
		err = commerr.ErrUnauthenticated

		return
	}

	uid = claims.UID
	userName = claims.UserName

	return
}
