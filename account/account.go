package account

import (
	// nolint: gosec
	"crypto/md5"
	"time"

	"github.com/sgostarter/i/commerr"
	"github.com/sgostarter/i/l"
	"github.com/sgostarter/libeasygo/crypt"
)

type Config struct {
	PasswordHashIterCount int    `yaml:"passwordHashIterCount" json:"passwordHashIterCount"`
	TokenSignKey          string `yaml:"tokenSignKey" json:"tokenSignKey"`

	TokenExpiresAfter time.Duration `yaml:"tokenExpiresAfter" json:"tokenExpiresAfter"`
	AutoRenewDuration time.Duration `yaml:"autoRenewDuration" json:"autoRenewDuration"`
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

	if cfg.TokenExpiresAfter <= 0 {
		cfg.TokenExpiresAfter = time.Hour * 24 * 356
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
	return impl.RegisterEx(0, accountName, password, nil)
}

func (impl *accountImpl) RegisterEx(userID uint64, accountName, password string, data []byte) (uid uint64, err error) {
	hashedPassword, err := crypt.HashPassword(password, impl.cfg.PasswordHashIterCount)
	if err != nil {
		return
	}

	uid, err = impl.storage.AddAccountEx(userID, accountName, hashedPassword, data)

	return
}

func (impl *accountImpl) Login(accountName, password string) (uid uint64, token string, err error) {
	uid, userHashedPassword, err := impl.storage.FindAccount(accountName)
	if err != nil {
		return
	}

	if userHashedPassword == "" {
		userHashedPassword, err = crypt.HashPassword(password, impl.cfg.PasswordHashIterCount)
		if err != nil {
			return
		}

		err = impl.storage.SetHashedPassword(uid, userHashedPassword)
		if err != nil {
			return
		}
	} else {
		ok := crypt.CheckHashedPassword(password, userHashedPassword, impl.cfg.PasswordHashIterCount)
		if !ok {
			err = commerr.ErrPermissionDenied

			return
		}
	}

	token, err = impl.tokenNew(uid, accountName)
	if err != nil {
		return
	}

	tokenExpiresAfter := impl.cfg.TokenExpiresAfter

	advanceConfig, err := impl.storage.GetAdvanceConfig(uid)
	if err != nil {
		return
	}

	if advanceConfig != nil && advanceConfig.TokenExpiresAfter > 0 {
		tokenExpiresAfter = advanceConfig.TokenExpiresAfter
	}

	err = impl.storage.AddToken(token, uid, time.Now().Add(tokenExpiresAfter))
	if err != nil {
		return
	}

	return
}

func (impl *accountImpl) RenameAccountName(uid uint64, newAccountName string) (err error) {
	err = impl.storage.RenameAccountName(uid, newAccountName)

	return
}

func (impl *accountImpl) SetAdvanceConfig(uid uint64, cfg *AdvanceConfig) (err error) {
	err = impl.storage.SetAdvanceConfig(uid, cfg)
	if err != nil {
		return
	}

	return
}

func (impl *accountImpl) GetAdvanceConfig(uid uint64) (cfg *AdvanceConfig, err error) {
	return impl.storage.GetAdvanceConfig(uid)
}

func (impl *accountImpl) Who(token string) (uid uint64, accountName string, err error) {
	return impl.who(token)
}

func (impl *accountImpl) GetData(uid uint64) (data []byte, err error) {
	return impl.storage.GetAccountData(uid)
}

func (impl *accountImpl) Logout(token string) error {
	return impl.storage.DelToken(token)
}

func (impl *accountImpl) HasAccount() (f bool, err error) {
	return impl.storage.HasAccount()
}

func (impl *accountImpl) ChangePassword(uid uint64, newPassword string) (err error) {
	hashedPassword, err := crypt.HashPassword(newPassword, impl.cfg.PasswordHashIterCount)
	if err != nil {
		return
	}

	err = impl.storage.SetHashedPassword(uid, hashedPassword)

	return
}

func (impl *accountImpl) ResetPassword(accountName string, newPassword string) (err error) {
	uid, exists, err := impl.storage.GetIDFromAccountName(accountName)
	if err != nil {
		return
	}

	if !exists {
		err = commerr.ErrNotFound

		return
	}

	hashedPassword, err := crypt.HashPassword(newPassword, impl.cfg.PasswordHashIterCount)
	if err != nil {
		return
	}

	err = impl.storage.SetHashedPassword(uid, hashedPassword)

	return
}

func (impl *accountImpl) ListUsers(createdAtStart, createdAtFinish int64) (accounts []User, err error) {
	return impl.storage.ListUsers(createdAtStart, createdAtFinish)
}

func (impl *accountImpl) SetPropertyData(token string, d interface{}) (err error) {
	_, userName, err := impl.who(token)
	if err != nil {
		return
	}

	return impl.storage.SetPropertyData(userName, d)
}

func (impl *accountImpl) SetPropertyDataByUserID(uid uint64, d interface{}) error {
	return impl.storage.SetPropertyDataByUserID(uid, d)
}

func (impl *accountImpl) GetPropertyData(token string, d interface{}) (err error) {
	_, userName, err := impl.who(token)
	if err != nil {
		return
	}

	return impl.storage.GetPropertyData(userName, d)
}

func (impl *accountImpl) GetPropertyDataByUserID(uid uint64, d interface{}) error {
	return impl.storage.GetPropertyDataByUserID(uid, d)
}

//
//
//

func (impl *accountImpl) who(token string) (uid uint64, accountName string, err error) {
	exists, err := impl.storage.TokenExists(token, impl.cfg.AutoRenewDuration)
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
