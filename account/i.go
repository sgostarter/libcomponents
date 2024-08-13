package account

import "time"

type User struct {
	UserName string
	UserID   uint64
	CreateAt int64
}

type AdvanceConfig struct {
	TokenExpiresAfter time.Duration `yaml:"tokenExpiresAfter" json:"tokenExpiresAfter"`
}

type Account interface {
	Register(accountName, password string) (uid uint64, err error)
	RegisterEx(userID uint64, accountName, password string, data []byte) (uid uint64, err error)
	Login(accountName, password string) (uid uint64, token string, err error)
	RenameAccountName(uid uint64, newAccountName string) (err error)
	SetAdvanceConfig(uid uint64, cfg *AdvanceConfig) error
	GetAdvanceConfig(uid uint64) (cfg *AdvanceConfig, err error)

	Who(token string) (uid uint64, accountName string, err error)
	GetData(uid uint64) (data []byte, err error)
	Logout(token string) error
	HasAccount() (f bool, err error)
	ChangePassword(uid uint64, newPassword string) (err error)
	ResetPassword(accountName string, newPassword string) (err error)
	ListUsers(createdAtStart, createdAtFinish int64) (accounts []User, err error)

	SetPropertyData(token string, d interface{}) error
	SetPropertyDataByUserID(uid uint64, d interface{}) error
	GetPropertyData(token string, d interface{}) error
	GetPropertyDataByUserID(uid uint64, d interface{}) error
}

type Storage interface {
	AddAccount(accountName, hashedPassword string) (uid uint64, err error)
	AddAccountEx(userID uint64, accountName, hashedPassword string, data []byte) (uid uint64, err error)
	SetHashedPassword(uid uint64, hashedPassword string) (err error)
	RenameAccountName(uid uint64, newAccountName string) error
	SetAdvanceConfig(uid uint64, cfg *AdvanceConfig) (err error)
	GetAdvanceConfig(uid uint64) (cfg *AdvanceConfig, err error)
	FindAccount(accountName string) (uid uint64, hashedPassword string, err error)
	GetAccountData(uid uint64) (data []byte, err error)
	HasAccount() (f bool, err error)
	ListUsers(createdAtStart, createdAtFinish int64) (accounts []User, err error)
	GetIDFromAccountName(accountName string) (uid uint64, exists bool, err error)

	AddToken(token string, uid uint64, expiredAt time.Time) error
	DelToken(token string) error
	TokenExists(token string, renewDuration time.Duration) (bool, error)

	SetPropertyData(accountName string, d interface{}) error
	SetPropertyDataByUserID(uid uint64, d interface{}) error
	GetPropertyData(accountName string, d interface{}) error
	GetPropertyDataByUserID(uid uint64, d interface{}) error
}
