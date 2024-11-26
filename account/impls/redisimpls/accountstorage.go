package redisimpls

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/godruoyi/go-snowflake"
	"github.com/sgostarter/i/commerr"
	"github.com/sgostarter/i/l"
	"github.com/sgostarter/libcomponents/account"
	"github.com/spf13/cast"
)

func NewRedisAccountStorage(preKey string, redisCli *redis.Client, logger l.Wrapper) account.Storage {
	if logger == nil {
		logger = l.NewNopLoggerWrapper()
	}

	logger = logger.WithFields(l.StringField(l.ClsKey, "accountsStorage"))

	if redisCli == nil {
		logger.Fatal("no redis client")
	}

	return &accountsStorage{
		logger:   logger,
		preKey:   preKey,
		redisCli: redisCli,
	}
}

type accountsStorage struct {
	logger   l.Wrapper
	preKey   string
	redisCli *redis.Client
}

func (impl *accountsStorage) AddAccount(accountName, hashedPassword string) (uid uint64, err error) {
	return impl.AddAccountEx(0, accountName, hashedPassword, nil)
}

func (impl *accountsStorage) AddAccountEx(userID uint64, accountName, hashedPassword string, data []byte) (uid uint64, err error) {
	uid = userID

	if uid == 0 {
		uid = snowflake.ID()
	}

	err = addAccountScript.Run(context.Background(), impl.redisCli, []string{impl.accountKey(uid),
		impl.accountNameKey(accountName), impl.accountCreateAtKey()}, uid, accountName, hashedPassword,
		data, time.Now().Unix()).Err()

	return
}

func (impl *accountsStorage) SetHashedPassword(uid uint64, hashedPassword string) (err error) {
	err = updateAccountPasswordScript.Run(context.Background(), impl.redisCli, []string{impl.accountKey(uid)},
		hashedPassword).Err()

	return
}

func (impl *accountsStorage) RenameAccountName(uid uint64, newAccountName string) error {
	return updateAccountNameScript.Run(context.Background(), impl.redisCli, []string{impl.accountKey(uid)},
		newAccountName).Err()
}

func (impl *accountsStorage) SetAdvanceConfig(uid uint64, cfg *account.AdvanceConfig) (err error) {
	var v string

	if cfg != nil {
		var vb []byte

		vb, err = json.Marshal(cfg)
		if err != nil {
			return
		}

		v = string(vb)
	}

	return updateAccountAdvanceConfigScript.Run(context.Background(), impl.redisCli, []string{impl.accountKey(uid)}, v).Err()
}

func (impl *accountsStorage) GetAdvanceConfig(uid uint64) (cfg *account.AdvanceConfig, err error) {
	d, err := impl.redisCli.HGet(context.Background(), impl.accountKey(uid), "adv_cfg").Bytes()
	if err != nil {
		return
	}

	cfg = new(account.AdvanceConfig)

	err = json.Unmarshal(d, cfg)

	return
}

func (impl *accountsStorage) FindAccount(accountName string) (uid uint64, hashedPassword string, err error) {
	uid, err = impl.redisCli.Get(context.Background(), impl.accountNameKey(accountName)).Uint64()
	if err != nil {
		return
	}

	hashedPassword, err = impl.redisCli.HGet(context.Background(), impl.accountKey(uid), "pass").Result()

	return
}

func (impl *accountsStorage) GetAccount(uid uint64) (accountName string, hashedPassword string, err error) {
	is, err := impl.redisCli.HMGet(context.Background(), impl.accountKey(uid), "name", "pass").Result()
	if err != nil {
		return
	}

	accountName, err = cast.ToStringE(is[0])
	if err != nil {
		return
	}

	hashedPassword, err = cast.ToStringE(is[1])

	return
}

func (impl *accountsStorage) GetAccountData(uid uint64) (data []byte, err error) {
	data, err = impl.redisCli.HGet(context.Background(), impl.accountKey(uid), "data").Bytes()

	return
}

func (impl *accountsStorage) HasAccount() (f bool, err error) {
	n, err := impl.redisCli.ZCard(context.Background(), impl.accountCreateAtKey()).Result()
	if err != nil {
		return
	}

	f = n > 0

	return
}

func (impl *accountsStorage) ListUsers(createdAtStart, createdAtFinish int64) (accounts []account.User, err error) {
	var minS, maxS string

	if createdAtStart <= 0 {
		minS = "-inf"
	} else {
		minS = strconv.FormatInt(createdAtStart, 10)
	}

	if createdAtFinish <= 0 {
		maxS = "+inf"
	} else {
		maxS = strconv.FormatInt(createdAtFinish, 10)
	}

	idSs, err := impl.redisCli.ZRangeByScore(context.Background(), impl.accountCreateAtKey(), &redis.ZRangeBy{
		Min: minS,
		Max: maxS,
	}).Result()
	if err != nil {
		return
	}

	accounts = make([]account.User, 0, len(idSs))

	for _, s := range idSs {
		id, e := strconv.ParseUint(s, 10, 64)
		if e != nil {
			impl.logger.WithFields(l.ErrorField(err), l.StringField("id", s)).
				Error("invalid uid on create_at table")

			continue
		}

		is, e := impl.redisCli.HMGet(context.Background(), impl.accountKey(id), "name", "create_at").Result()
		if e != nil {
			if errors.Is(e, redis.Nil) {
				continue
			}

			err = e

			return
		}

		userName := cast.ToString(is[0])
		creteAt := cast.ToInt64(is[1])

		accounts = append(accounts, account.User{
			UserName: userName,
			UserID:   id,
			CreateAt: creteAt,
		})
	}

	return
}

func (impl *accountsStorage) GetIDFromAccountName(accountName string) (uid uint64, exists bool, err error) {
	uid, err = impl.redisCli.Get(context.Background(), impl.accountNameKey(accountName)).Uint64()
	if err == nil {
		exists = true

		return
	}

	if errors.Is(err, redis.Nil) {
		err = nil

		return
	}

	return
}

func (impl *accountsStorage) AddToken(token string, uid uint64, expiredAt time.Time) error {
	d := time.Until(expiredAt)
	if d <= 0 {
		return nil
	}

	return tokenAddScript.Run(context.Background(), impl.redisCli, []string{impl.accountTokenKey(token),
		impl.accountIdTokensKey(uid)}, token, uid, int64(d.Seconds())).Err()
}

func (impl *accountsStorage) DelToken(token string) (err error) {
	uid, err := impl.redisCli.Get(context.Background(), impl.accountTokenKey(token)).Uint64()
	if err != nil {
		return
	}

	err = tokenDelScript.Run(context.Background(), impl.redisCli, []string{impl.accountTokenKey(token),
		impl.accountIdTokensKey(uid)}, token).Err()

	return
}

func (impl *accountsStorage) TokenExists(token string, renewDuration time.Duration) (exists bool, err error) {
	seconds := int64(renewDuration.Seconds())

	exists, err = tokenCheckScript.Run(context.Background(), impl.redisCli, []string{impl.accountTokenKey(token)},
		seconds).Bool()

	return
}

func (impl *accountsStorage) SetPropertyData(accountName string, d interface{}) (err error) {
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

func (impl *accountsStorage) SetPropertyDataByUserID(uid uint64, data interface{}) (err error) {
	d, err := json.Marshal(data)
	if err != nil {
		return
	}

	return updateAccountPropertyDataScript.Run(context.Background(), impl.redisCli, []string{impl.accountKey(uid)},
		d).Err()
}

func (impl *accountsStorage) GetPropertyData(accountName string, d interface{}) (err error) {
	uid, exists, err := impl.GetIDFromAccountName(accountName)
	if err != nil {
		return
	}

	if !exists {
		err = commerr.ErrNotFound

		return
	}

	return impl.GetPropertyDataByUserID(uid, d)
}

func (impl *accountsStorage) GetPropertyDataByUserID(uid uint64, d interface{}) (err error) {
	ds, err := impl.redisCli.HGet(context.Background(), impl.accountKey(uid), "property_data").Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			err = nil
		}

		return
	}

	err = json.Unmarshal(ds, d)

	return
}

//
//
//

func (impl *accountsStorage) accountKey(userID uint64) string {
	return impl.preKey + "uid:" + strconv.FormatUint(userID, 10)
}

func (impl *accountsStorage) accountNameKey(userName string) string {
	return impl.preKey + "un:" + userName
}

func (impl *accountsStorage) accountCreateAtKey() string {
	return impl.preKey + "users:create_at"
}

func (impl *accountsStorage) accountTokenKey(token string) string {
	return impl.preKey + "utk:" + token
}

func (impl *accountsStorage) accountIdTokensKey(userID uint64) string {
	return impl.preKey + "utk-s:" + strconv.FormatUint(userID, 10)
}
