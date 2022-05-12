package wallet

import (
	"context"
	"errors"
	"time"

	"github.com/go-redis/redis/v8"
)

var (
	ErrInvalidObject = errors.New("invalidObject")
)

func NewRedisWallet(redisCli *redis.Client, redisKeyPre string) Wallet {
	history := newRedisHistory(redisCli, redisKeyPre)

	return &redisWalletImpl{
		history:     history,
		redisCli:    redisCli,
		redisKeyPre: redisKeyPre,
	}
}

type redisWalletImpl struct {
	history     *redisHistoryImpl
	redisCli    *redis.Client
	redisKeyPre string
}

func (impl *redisWalletImpl) GetHistory() History {
	return impl.history
}

func (impl *redisWalletImpl) TransToLocker(ctx context.Context, account string, coins int64, remark string, locker Locker, toAccount, key string, options ...Option) (err error) {
	flag, err := optionNew(options...).ConflictFlag()
	if err != nil {
		return err
	}

	rLocker, ok := locker.(*redisLockerImpl)
	if !ok {
		err = ErrInvalidObject

		return
	}

	val, err := walletTrans2LockerScript.Run(ctx, impl.redisCli, []string{impl.walletRedisKey(), rLocker.accountRedisKey(toAccount), impl.history.accountRedisKey(account)},
		account, coins, key, totalKey, flag, time.Now().Unix(), BuildHistoryValuePayload(toAccount, key, remark)).Int()
	if err != nil {
		return err
	}

	if val == 0 {
		return
	}

	switch val {
	case 1:
		err = ErrNoCoins
	case 2:
		err = ErrExists
	default:
		err = ErrFailed
	}

	return err
}

func (impl *redisWalletImpl) GetCoins(ctx context.Context, account string) (val int64, err error) {
	val, err = impl.redisCli.HGet(ctx, impl.walletRedisKey(), account).Int64()
	if errors.Is(err, redis.Nil) {
		err = nil
	}

	return
}

func (impl *redisWalletImpl) walletRedisKey() string {
	if impl.redisKeyPre == "" {
		return "wallet"
	}

	return impl.redisKeyPre + ":" + "wallet"
}
