package account

import (
	// nolint: gosec
	"crypto/md5"

	"github.com/sgostarter/i/commerr"
	"github.com/sgostarter/i/l"
	"github.com/sgostarter/libeasygo/crypt"
)

type Config struct {
	PasswordHashIterCount int
	TokenSignKey          string
}

func NewAccount(storage Storage, cfg *Config, logger l.Wrapper) Account {
	if logger == nil {
		logger = l.NewNopLoggerWrapper()
	}

	if storage == nil {
		logger.Error("no storage")

		return nil
	}

	if cfg == nil {
		logger.Error("no config")

		return nil
	}

	if cfg.PasswordHashIterCount <= 0 {
		cfg.PasswordHashIterCount = 4096
	}

	tokenKey := md5.Sum([]byte(cfg.TokenSignKey)) // nolint: gosec

	return &accountImpl{
		logger:   logger.WithFields(l.StringField(l.ClsKey, "accountImpl")),
		storage:  storage,
		cfg:      cfg,
		tokenKey: tokenKey[:],
	}
}

type accountImpl struct {
	logger  l.Wrapper
	storage Storage
	cfg     *Config

	tokenKey []byte
}

func (impl *accountImpl) Register(accountName, password string) (uid uint64, err error) {
	hashedPassword, err := crypt.HashPassword(password, impl.cfg.PasswordHashIterCount)
	if err != nil {
		return
	}

	uid, err = impl.storage.AddAccount(accountName, hashedPassword)

	return
}

func (impl *accountImpl) Login(accountName, password string) (uid uint64, token string, err error) {
	hashedPassword, err := crypt.HashPassword(password, impl.cfg.PasswordHashIterCount)
	if err != nil {
		return
	}

	uid, userHashedPassword, err := impl.storage.FindAccount(accountName)
	if err != nil {
		return
	}

	if hashedPassword != userHashedPassword {
		err = commerr.ErrPermissionDenied

		return
	}

	token, err = impl.tokenNew(uid, accountName)
	if err != nil {
		return
	}

	err = impl.storage.AddToken(token)
	if err != nil {
		return
	}

	return
}

func (impl *accountImpl) Who(token string) (uid uint64, accountName string, err error) {
	return impl.who(token)
}

func (impl *accountImpl) Logout(token string) error {
	return impl.storage.DelToken(token)
}

func (impl *accountImpl) SetPropertyData(token string, d interface{}) (err error) {
	_, userName, err := impl.who(token)
	if err != nil {
		return
	}

	return impl.storage.SetPropertyData(userName, d)
}

func (impl *accountImpl) GetPropertyData(token string, d interface{}) (err error) {
	_, userName, err := impl.who(token)
	if err != nil {
		return
	}

	return impl.storage.GetPropertyData(userName, d)
}

//
//
//

func (impl *accountImpl) who(token string) (uid uint64, accountName string, err error) {
	exists, err := impl.storage.TokenExists(token)
	if err != nil {
		return
	}

	if !exists {
		err = commerr.ErrNotFound

		return
	}

	uid, accountName, err = impl.tokenCheck(token)

	return
}
