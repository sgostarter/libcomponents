package account

import "time"

type User struct {
	UserName string
	UserID   uint64
	CreateAt int64
}

type Account interface {
	Register(accountName, password string) (uid uint64, err error)
	RegisterEx(userID uint64, accountName, password string) (uid uint64, err error)
	Login(accountName, password string) (uid uint64, token string, err error)
	Who(token string) (uid uint64, accountName string, err error)
	Logout(token string) error
	HasAccount() (f bool, err error)
	ChangePassword(token string, newPassword string) (err error)
	ResetPassword(accountName string, newPassword string) (err error)
	ListUsers(createdAtStart, createdAtFinish int64) (accounts []User, err error)
	UserID2Name(uid uint64) (userName string, err error)

	SetPropertyData(token string, d interface{}) error
	SetPropertyDataByUserID(uid uint64, d interface{}) error
	GetPropertyData(token string, d interface{}) error
	GetPropertyDataByUserID(uid uint64, d interface{}) error
}

type Storage interface {
	AddAccount(accountName, hashedPassword string) (uid uint64, err error)
	AddAccountEx(userID uint64, accountName, hashedPassword string) (uid uint64, err error)
	SetHashedPassword(accountName, hashedPassword string) (err error)
	FindAccount(accountName string) (uid uint64, hashedPassword string, err error)
	HasAccount() (f bool, err error)
	ListUsers(createdAtStart, createdAtFinish int64) (accounts []User, err error)
	UserID2Name(uid uint64) (userName string, err error)

	AddToken(token string, expiredAt time.Time) error
	DelToken(token string) error
	TokenExists(token string, renewDuration time.Duration) (bool, error)

	SetPropertyData(accountName string, d interface{}) error
	SetPropertyDataByUserID(uid uint64, d interface{}) error
	GetPropertyData(accountName string, d interface{}) error
	GetPropertyDataByUserID(uid uint64, d interface{}) error
}
