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
	"github.com/spf13/cast"
)

func NewFMAccountStorage(root string, storage stg.FileStorage) account.Storage {
	return NewFMAccountStorageEx(root, storage, false)
}

func NewFMAccountStorageEx(root string, storage stg.FileStorage, prettySerial bool) account.Storage {
	if storage == nil {
		storage = rawfs.NewFSStorage("")
	}

	impl := &fsAccountStorageImpl{
		accountStorage: mwf.NewMemWithFile[map[uint64]*AccountInfo, mwf.Serial, mwf.Lock](
			make(map[uint64]*AccountInfo), &mwf.JSONSerial{
				MarshalIndent: prettySerial,
			}, &sync.RWMutex{}, filepath.Join(root, "accounts.json"), storage),
		tokenStorage: mwf.NewMemWithFile[map[string]*TokenInfo, mwf.Serial, mwf.Lock](
			make(map[string]*TokenInfo), &mwf.JSONSerial{
				MarshalIndent: prettySerial,
			}, &sync.RWMutex{}, filepath.Join(root, "tokens.json"), storage),
		accountPropertyStorage: mwf.NewMemWithFile[map[uint64][]byte, mwf.Serial, mwf.Lock](
			make(map[uint64][]byte), &mwf.JSONSerial{
				MarshalIndent: prettySerial,
			}, &sync.RWMutex{}, filepath.Join(root, "accountProperties.json"), storage),
	}

	impl.init()

	return impl
}

type TokenInfo struct {
	ExpiredAt time.Time
	UID       uint64
}

type fsAccountStorageImpl struct {
	accountStorage          *mwf.MemWithFile[map[uint64]*AccountInfo, mwf.Serial, mwf.Lock] // uid -> user info
	tokenStorage            *mwf.MemWithFile[map[string]*TokenInfo, mwf.Serial, mwf.Lock]   // token -> uid, expireAt
	accountPropertyStorage  *mwf.MemWithFile[map[uint64][]byte, mwf.Serial, mwf.Lock]       // uid -> property
	lastCleanExpiredTokenAt time.Time

	accountName2UserID sync.Map // account name -> uid
}

func (impl *fsAccountStorageImpl) init() {
	_ = impl.accountStorage.Change(func(oldM map[uint64]*AccountInfo) (newM map[uint64]*AccountInfo, err error) {
		newM = oldM
		if len(newM) == 0 {
			newM = make(map[uint64]*AccountInfo)
		}

		var changed bool

		for _, info := range newM {
			impl.accountName2UserID.Store(info.AccountName, info.ID)

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
	return impl.AddAccountEx(0, accountName, hashedPassword, nil)
}

func (impl *fsAccountStorageImpl) GetIDFromAccountName(accountName string) (uid uint64, exists bool, err error) {
	i, ok := impl.accountName2UserID.Load(accountName)
	if !ok {
		return
	}

	uid = cast.ToUint64(i)
	exists = true

	return
}

func (impl *fsAccountStorageImpl) AddAccountEx(userID uint64, accountName, hashedPassword string, data []byte) (uid uint64, err error) {
	err = impl.accountStorage.Change(func(oldM map[uint64]*AccountInfo) (newM map[uint64]*AccountInfo, err error) {
		newM = oldM
		if len(newM) == 0 {
			newM = make(map[uint64]*AccountInfo)
		}

		_, exists, _ := impl.GetIDFromAccountName(accountName)
		if exists {
			err = commerr.ErrAlreadyExists

			return
		}

		uid = userID

		if uid == 0 {
			uid = snowflake.ID()
		}

		_, exists = newM[uid]
		if exists {
			err = commerr.ErrAlreadyExists

			return
		}

		newM[uid] = &AccountInfo{
			ID:             uid,
			AccountName:    accountName,
			HashedPassword: hashedPassword,
			CreateAt:       time.Now().Unix(),
			Data:           data,
		}

		impl.accountName2UserID.Store(accountName, uid)

		return
	})

	return
}

func (impl *fsAccountStorageImpl) SetHashedPassword(uid uint64, hashedPassword string) (err error) {
	err = impl.accountStorage.Change(func(oldM map[uint64]*AccountInfo) (newM map[uint64]*AccountInfo, err error) {
		newM = oldM
		if len(newM) == 0 {
			newM = make(map[uint64]*AccountInfo)
		}

		if _, ok := newM[uid]; !ok {
			err = commerr.ErrNotFound

			return
		}

		newM[uid].HashedPassword = hashedPassword

		return
	})

	return
}

func (impl *fsAccountStorageImpl) SetAdvanceConfig(uid uint64, cfg *account.AdvanceConfig) (err error) {
	err = impl.accountStorage.Change(func(oldM map[uint64]*AccountInfo) (newM map[uint64]*AccountInfo, err error) {
		newM = oldM
		if len(newM) == 0 {
			newM = make(map[uint64]*AccountInfo)
		}

		if _, ok := newM[uid]; !ok {
			err = commerr.ErrNotFound

			return
		}

		newM[uid].Cfg = cfg

		return
	})

	return
}

func (impl *fsAccountStorageImpl) GetAdvanceConfig(uid uint64) (cfg *account.AdvanceConfig, err error) {
	err = impl.accountStorage.Change(func(oldM map[uint64]*AccountInfo) (newM map[uint64]*AccountInfo, err error) {
		newM = oldM
		if len(newM) == 0 {
			newM = make(map[uint64]*AccountInfo)
		}

		if _, ok := newM[uid]; !ok {
			err = commerr.ErrNotFound

			return
		}

		cfg = newM[uid].Cfg

		return
	})

	return
}

func (impl *fsAccountStorageImpl) FindAccount(accountName string) (uid uint64, hashedPassword string, err error) {
	uid, exists, err := impl.GetIDFromAccountName(accountName)
	if err != nil {
		return
	}

	if !exists {
		err = commerr.ErrNotFound

		return
	}

	impl.accountStorage.Read(func(m map[uint64]*AccountInfo) {
		if info, ok := m[uid]; ok {
			uid = info.ID
			hashedPassword = info.HashedPassword
		} else {
			err = commerr.ErrNotFound
		}
	})

	return
}

func (impl *fsAccountStorageImpl) GetAccount(uid uint64) (accountName string, hashedPassword string, err error) {
	impl.accountStorage.Read(func(m map[uint64]*AccountInfo) {
		if info, ok := m[uid]; ok {
			accountName = info.AccountName
			hashedPassword = info.HashedPassword
		} else {
			err = commerr.ErrNotFound
		}
	})

	return
}

func (impl *fsAccountStorageImpl) GetAccountData(uid uint64) (data []byte, err error) {
	impl.accountStorage.Read(func(m map[uint64]*AccountInfo) {
		if info, ok := m[uid]; ok {
			data = info.Data
		} else {
			err = commerr.ErrNotFound
		}
	})

	return
}

func (impl *fsAccountStorageImpl) HasAccount() (f bool, err error) {
	impl.accountStorage.Read(func(m map[uint64]*AccountInfo) {
		f = len(m) > 0
	})

	return
}

func (impl *fsAccountStorageImpl) ListUsers(createdAtStart, createdAtFinish int64) (accounts []account.User, err error) {
	impl.accountStorage.Read(func(m map[uint64]*AccountInfo) {
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

func (impl *fsAccountStorageImpl) RenameAccountName(uid uint64, newAccountName string) error {
	return impl.accountStorage.Change(func(oldM map[uint64]*AccountInfo) (newM map[uint64]*AccountInfo, err error) {
		newM = oldM
		if len(newM) == 0 {
			newM = make(map[uint64]*AccountInfo)
		}

		ai, ok := newM[uid]
		if !ok {
			err = commerr.ErrNotFound

			return
		}

		if ai.AccountName == newAccountName {
			return
		}

		_, exists, err := impl.GetIDFromAccountName(newAccountName)
		if err != nil {
			return
		}

		if exists {
			err = commerr.ErrAlreadyExists

			return
		}

		impl.accountName2UserID.Delete(ai.AccountName)
		impl.accountName2UserID.Store(newAccountName, uid)

		newM[uid].AccountName = newAccountName

		return
	})
}

func (impl *fsAccountStorageImpl) cleanExpiredTokenOnSafe(m map[string]*TokenInfo) (cleanedCount int) {
	impl.lastCleanExpiredTokenAt = time.Now()

	for k, tokenInfo := range m {
		if time.Now().After(tokenInfo.ExpiredAt) {
			delete(m, k)

			cleanedCount++
		}
	}

	return
}

func (impl *fsAccountStorageImpl) AddToken(token string, uid uint64, expiredAt time.Time) error {
	return impl.tokenStorage.Change(func(oldM map[string]*TokenInfo) (newM map[string]*TokenInfo, err error) {
		newM = oldM
		if len(newM) == 0 {
			newM = make(map[string]*TokenInfo)
		}

		if _, ok := newM[token]; ok {
			err = commerr.ErrAlreadyExists

			return
		}

		impl.cleanExpiredTokenOnSafe(newM)

		newM[token] = &TokenInfo{
			ExpiredAt: expiredAt,
			UID:       uid,
		}

		return
	})
}

func (impl *fsAccountStorageImpl) DelToken(token string) error {
	return impl.tokenStorage.Change(func(oldM map[string]*TokenInfo) (newM map[string]*TokenInfo, err error) {
		newM = oldM
		if len(newM) == 0 {
			newM = make(map[string]*TokenInfo)
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
	impl.tokenStorage.Read(func(m map[string]*TokenInfo) {
		_, exists = m[token]
	})

	if exists && renewDuration > 0 {
		_ = impl.tokenStorage.Change(func(oldM map[string]*TokenInfo) (newM map[string]*TokenInfo, err error) {
			newM = oldM
			if len(newM) == 0 {
				newM = make(map[string]*TokenInfo)
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

			newM[token].ExpiredAt = expiredAt

			return
		})
	}

	if time.Since(impl.lastCleanExpiredTokenAt) > time.Hour {
		_ = impl.tokenStorage.Change(func(oldM map[string]*TokenInfo) (newM map[string]*TokenInfo, err error) {
			newM = oldM
			if len(newM) == 0 {
				newM = make(map[string]*TokenInfo)
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

func (impl *fsAccountStorageImpl) SetPropertyData(accountName string, d interface{}) (err error) {
	uid, exists, err := impl.GetIDFromAccountName(accountName)
	if err != nil {
		return
	}

	if !exists {
		err = commerr.ErrNotFound

		return
	}

	return impl.SetPropertyDataByUserID(uid, d)
}

func (impl *fsAccountStorageImpl) SetPropertyDataByUserID(uid uint64, d interface{}) error {
	dd, err := json.Marshal(d)
	if err != nil {
		return err
	}

	return impl.accountPropertyStorage.Change(func(oldM map[uint64][]byte) (newM map[uint64][]byte, err error) {
		newM = oldM
		if len(newM) == 0 {
			newM = make(map[uint64][]byte)
		}

		newM[uid] = dd

		return
	})
}

func (impl *fsAccountStorageImpl) GetPropertyData(accountName string, d interface{}) error {
	uid, exists, err := impl.GetIDFromAccountName(accountName)
	if err != nil {
		return err
	}

	if !exists {
		return commerr.ErrNotFound
	}

	return impl.GetPropertyDataByUserID(uid, d)
}

func (impl *fsAccountStorageImpl) GetPropertyDataByUserID(uid uint64, d interface{}) (err error) {
	impl.accountPropertyStorage.Read(func(m map[uint64][]byte) {
		if dd, ok := m[uid]; ok {
			err = json.Unmarshal(dd, d)
		} else {
			err = commerr.ErrNotFound
		}
	})

	return
}
