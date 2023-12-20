package account

import "time"

type Account interface {
	Register(accountName, password string) (uid uint64, err error)
	Login(accountName, password string) (uid uint64, token string, err error)
	Who(token string) (uid uint64, accountName string, err error)
	Logout(token string) error

	SetPropertyData(token string, d interface{}) error
	GetPropertyData(token string, d interface{}) error
}

type Storage interface {
	AddAccount(accountName, hashedPassword string) (uid uint64, err error)
	SetHashedPassword(accountName, hashedPassword string) (err error)
	FindAccount(accountName string) (uid uint64, hashedPassword string, err error)

	AddToken(token string, expiredAt time.Time) error
	DelToken(token string) error
	TokenExists(token string, renewDuration time.Duration) (bool, error)

	SetPropertyData(accountName string, d interface{}) error
	GetPropertyData(accountName string, d interface{}) error
}
