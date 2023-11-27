package fmaccountstorage

import (
	"encoding/json"
	"path/filepath"
	"sync"
	"time"

	"github.com/godruoyi/go-snowflake"
	"github.com/sgostarter/i/commerr"
	"github.com/sgostarter/i/stg"
	"github.com/sgostarter/libcomponents/account"
	"github.com/sgostarter/libeasygo/stg/fs/rawfs"
	"github.com/sgostarter/libeasygo/stg/mwf"
)

func NewFMAccountStorage(root string, storage stg.FileStorage) account.Storage {
	if storage == nil {
		storage = rawfs.NewFSStorage("")
	}

	return &fsAccountStorageImpl{
		accountStorage: mwf.NewMemWithFile[map[string]*AccountInfo, mwf.Serial, mwf.Lock](
			make(map[string]*AccountInfo), &mwf.JSONSerial{}, &sync.RWMutex{}, filepath.Join(root, "accounts.json"), storage),
		tokenStorage: mwf.NewMemWithFile[map[string]time.Time, mwf.Serial, mwf.Lock](
			make(map[string]time.Time), &mwf.JSONSerial{}, &sync.RWMutex{}, filepath.Join(root, "tokens.json"), storage),
		accountPropertyStorage: mwf.NewMemWithFile[map[string][]byte, mwf.Serial, mwf.Lock](
			make(map[string][]byte), &mwf.JSONSerial{}, &sync.RWMutex{}, filepath.Join(root, "accountProperties.json"), storage),
	}
}

type fsAccountStorageImpl struct {
	accountStorage         *mwf.MemWithFile[map[string]*AccountInfo, mwf.Serial, mwf.Lock]
	tokenStorage           *mwf.MemWithFile[map[string]time.Time, mwf.Serial, mwf.Lock]
	accountPropertyStorage *mwf.MemWithFile[map[string][]byte, mwf.Serial, mwf.Lock]
}

func (impl *fsAccountStorageImpl) AddAccount(accountName, hashedPassword string) (uid uint64, err error) {
	err = impl.accountStorage.Change(func(oldM map[string]*AccountInfo) (newM map[string]*AccountInfo, err error) {
		newM = oldM
		if len(newM) == 0 {
			newM = make(map[string]*AccountInfo)
		}

		if _, ok := newM[accountName]; ok {
			err = commerr.ErrExiting

			return
		}

		uid = snowflake.ID()
		newM[accountName] = &AccountInfo{
			ID:             uid,
			AccountName:    accountName,
			HashedPassword: hashedPassword,
		}

		return
	})

	return
}

func (impl *fsAccountStorageImpl) FindAccount(accountName string) (uid uint64, hashedPassword string, err error) {
	impl.accountStorage.Read(func(m map[string]*AccountInfo) {
		if info, ok := m[accountName]; ok {
			uid = info.ID
			hashedPassword = info.HashedPassword
		} else {
			err = commerr.ErrNotFound
		}
	})

	return
}

func (impl *fsAccountStorageImpl) AddToken(token string) error {
	return impl.tokenStorage.Change(func(oldM map[string]time.Time) (newM map[string]time.Time, err error) {
		newM = oldM
		if len(newM) == 0 {
			newM = make(map[string]time.Time)
		}

		if _, ok := newM[token]; ok {
			err = commerr.ErrAlreadyExists

			return
		}

		newM[token] = time.Now()

		return
	})
}

func (impl *fsAccountStorageImpl) DelToken(token string) error {
	return impl.tokenStorage.Change(func(oldM map[string]time.Time) (newM map[string]time.Time, err error) {
		newM = oldM
		if len(newM) == 0 {
			newM = make(map[string]time.Time)
		}

		if _, ok := newM[token]; !ok {
			err = commerr.ErrNotFound

			return
		}

		delete(newM, token)

		return
	})
}

func (impl *fsAccountStorageImpl) TokenExists(token string) (exists bool, err error) {
	impl.tokenStorage.Read(func(m map[string]time.Time) {
		_, exists = m[token]
	})

	return
}

func (impl *fsAccountStorageImpl) SetPropertyData(accountName string, d interface{}) error {
	dd, err := json.Marshal(d)
	if err != nil {
		return err
	}

	return impl.accountPropertyStorage.Change(func(oldM map[string][]byte) (newM map[string][]byte, err error) {
		newM = oldM
		if len(newM) == 0 {
			newM = make(map[string][]byte)
		}

		newM[accountName] = dd

		return
	})
}

func (impl *fsAccountStorageImpl) GetPropertyData(accountName string, d interface{}) (err error) {
	impl.accountPropertyStorage.Read(func(m map[string][]byte) {
		if dd, ok := m[accountName]; ok {
			err = json.Unmarshal(dd, d)
		} else {
			err = commerr.ErrNotFound
		}
	})

	return
}
