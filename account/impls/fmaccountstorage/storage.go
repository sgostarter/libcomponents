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
	return NewFMAccountStorageEx(root, storage, false)
}

func NewFMAccountStorageEx(root string, storage stg.FileStorage, prettySerial bool) account.Storage {
	if storage == nil {
		storage = rawfs.NewFSStorage("")
	}

	impl := &fsAccountStorageImpl{
		accountStorage: mwf.NewMemWithFile[map[string]*AccountInfo, mwf.Serial, mwf.Lock](
			make(map[string]*AccountInfo), &mwf.JSONSerial{
				MarshalIndent: prettySerial,
			}, &sync.RWMutex{}, filepath.Join(root, "accounts.json"), storage),
		tokenStorage: mwf.NewMemWithFile[map[string]time.Time, mwf.Serial, mwf.Lock](
			make(map[string]time.Time), &mwf.JSONSerial{
				MarshalIndent: prettySerial,
			}, &sync.RWMutex{}, filepath.Join(root, "tokens.json"), storage),
		accountPropertyStorage: mwf.NewMemWithFile[map[string][]byte, mwf.Serial, mwf.Lock](
			make(map[string][]byte), &mwf.JSONSerial{
				MarshalIndent: prettySerial,
			}, &sync.RWMutex{}, filepath.Join(root, "accountProperties.json"), storage),
	}

	impl.init()

	return impl
}

type fsAccountStorageImpl struct {
	accountStorage          *mwf.MemWithFile[map[string]*AccountInfo, mwf.Serial, mwf.Lock]
	tokenStorage            *mwf.MemWithFile[map[string]time.Time, mwf.Serial, mwf.Lock]
	accountPropertyStorage  *mwf.MemWithFile[map[string][]byte, mwf.Serial, mwf.Lock]
	lastCleanExpiredTokenAt time.Time

	userID2Name sync.Map
}

func (impl *fsAccountStorageImpl) init() {
	_ = impl.accountStorage.Change(func(oldM map[string]*AccountInfo) (newM map[string]*AccountInfo, err error) {
		newM = oldM
		if len(newM) == 0 {
			newM = make(map[string]*AccountInfo)
		}

		var changed bool

		for _, info := range newM {
			impl.userID2Name.Store(info.ID, info.AccountName)

			if info.CreateAt == 0 {
				info.CreateAt = time.Now().Unix()

				changed = true
			}
		}

		if !changed {
			err = commerr.ErrAborted

			return
		}

		return
	})
}

func (impl *fsAccountStorageImpl) AddAccount(accountName, hashedPassword string) (uid uint64, err error) {
	return impl.AddAccountEx(0, accountName, hashedPassword)
}

func (impl *fsAccountStorageImpl) AddAccountEx(userID uint64, accountName, hashedPassword string) (uid uint64, err error) {
	err = impl.accountStorage.Change(func(oldM map[string]*AccountInfo) (newM map[string]*AccountInfo, err error) {
		newM = oldM
		if len(newM) == 0 {
			newM = make(map[string]*AccountInfo)
		}

		if _, ok := newM[accountName]; ok {
			err = commerr.ErrAlreadyExists

			return
		}

		uid = userID

		if uid == 0 {
			uid = snowflake.ID()
		} else {
			for _, info := range newM {
				if info.ID == uid {
					err = commerr.ErrAlreadyExists

					return
				}
			}
		}

		newM[accountName] = &AccountInfo{
			ID:             uid,
			AccountName:    accountName,
			HashedPassword: hashedPassword,
			CreateAt:       time.Now().Unix(),
		}

		impl.userID2Name.Store(uid, accountName)

		return
	})

	return
}

func (impl *fsAccountStorageImpl) SetHashedPassword(accountName, hashedPassword string) (err error) {
	err = impl.accountStorage.Change(func(oldM map[string]*AccountInfo) (newM map[string]*AccountInfo, err error) {
		newM = oldM
		if len(newM) == 0 {
			newM = make(map[string]*AccountInfo)
		}

		if _, ok := newM[accountName]; !ok {
			err = commerr.ErrNotFound

			return
		}

		newM[accountName].HashedPassword = hashedPassword

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

func (impl *fsAccountStorageImpl) HasAccount() (f bool, err error) {
	impl.accountStorage.Read(func(m map[string]*AccountInfo) {
		f = len(m) > 0
	})

	return
}

func (impl *fsAccountStorageImpl) ListUsers(createdAtStart, createdAtFinish int64) (accounts []account.User, err error) {
	impl.accountStorage.Read(func(m map[string]*AccountInfo) {
		accounts = make([]account.User, 0, len(m))

		for _, info := range m {
			if (createdAtStart > 0 && info.CreateAt < createdAtStart) ||
				(createdAtFinish > 0 && info.CreateAt > createdAtFinish) {
				continue
			}

			accounts = append(accounts, account.User{
				UserName: info.AccountName,
				UserID:   info.ID,
				CreateAt: info.CreateAt,
			})
		}
	})

	return
}

func (impl *fsAccountStorageImpl) UserID2Name(uid uint64) (userName string, err error) {
	i, ok := impl.userID2Name.Load(uid)
	if !ok {
		err = commerr.ErrNotFound

		return
	}

	userName, ok = i.(string)
	if !ok {
		err = commerr.ErrInternal

		return
	}

	return
}

func (impl *fsAccountStorageImpl) cleanExpiredTokenOnSafe(m map[string]time.Time) (cleanedCount int) {
	impl.lastCleanExpiredTokenAt = time.Now()

	for k, expiredAt := range m {
		if time.Now().After(expiredAt) {
			delete(m, k)

			cleanedCount++
		}
	}

	return
}

func (impl *fsAccountStorageImpl) AddToken(token string, expiredAt time.Time) error {
	return impl.tokenStorage.Change(func(oldM map[string]time.Time) (newM map[string]time.Time, err error) {
		newM = oldM
		if len(newM) == 0 {
			newM = make(map[string]time.Time)
		}

		if _, ok := newM[token]; ok {
			err = commerr.ErrAlreadyExists

			return
		}

		impl.cleanExpiredTokenOnSafe(newM)

		newM[token] = expiredAt

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

		impl.cleanExpiredTokenOnSafe(newM)

		delete(newM, token)

		return
	})
}

func (impl *fsAccountStorageImpl) TokenExists(token string, renewDuration time.Duration) (exists bool, err error) {
	impl.tokenStorage.Read(func(m map[string]time.Time) {
		_, exists = m[token]
	})

	if exists && renewDuration > 0 {
		_ = impl.tokenStorage.Change(func(oldM map[string]time.Time) (newM map[string]time.Time, err error) {
			newM = oldM
			if len(newM) == 0 {
				newM = make(map[string]time.Time)
			}

			var i any

			i, exists = newM[token]
			if !exists {
				err = commerr.ErrNotFound

				return
			}

			expiredAt, ok := i.(time.Time)
			if !ok {
				err = commerr.ErrInternal

				return
			}

			expiredAt = expiredAt.Add(renewDuration)

			newM[token] = expiredAt

			return
		})
	}

	if time.Since(impl.lastCleanExpiredTokenAt) > time.Hour {
		_ = impl.tokenStorage.Change(func(oldM map[string]time.Time) (newM map[string]time.Time, err error) {
			newM = oldM
			if len(newM) == 0 {
				newM = make(map[string]time.Time)
			}

			if impl.cleanExpiredTokenOnSafe(newM) <= 0 {
				err = commerr.ErrAborted

				return
			}

			return
		})
	}

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

func (impl *fsAccountStorageImpl) SetPropertyDataByUserID(uid uint64, d interface{}) error {
	i, ok := impl.userID2Name.Load(uid)
	if !ok {
		return commerr.ErrNotFound
	}

	userName, ok := i.(string)
	if !ok {
		return commerr.ErrInternal
	}

	return impl.SetPropertyData(userName, d)
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

func (impl *fsAccountStorageImpl) GetPropertyDataByUserID(uid uint64, d interface{}) error {
	i, ok := impl.userID2Name.Load(uid)
	if !ok {
		return commerr.ErrNotFound
	}

	userName, ok := i.(string)
	if !ok {
		return commerr.ErrInternal
	}

	return impl.GetPropertyData(userName, d)
}
