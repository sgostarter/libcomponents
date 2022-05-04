package wallet

import (
	"context"
	"errors"
	"time"

	"github.com/go-redis/redis/v8"
)

const (
	totalKey = "total"
)

var (
	ErrConflict = errors.New("conflict")
)

func NewRedisLocker(redisCli *redis.Client, redisKeyPre string) Locker {
	return &redisLockerImpl{
		redisCli:    redisCli,
		redisKeyPre: redisKeyPre,
	}
}

type redisLockerImpl struct {
	redisCli    *redis.Client
	redisKeyPre string
}

func (impl *redisLockerImpl) accountRedisKey(account string) string {
	redisKey := "locker:" + account
	if impl.redisKeyPre != "" {
		redisKey = impl.redisKeyPre + ":" + redisKey
	}

	return redisKey
}

func (impl *redisLockerImpl) Set(ctx context.Context, account, key string, coins int64, options ...Option) error {
	flag, err := optionNew(options...).ConflictFlag()
	if err != nil {
		return err
	}

	return lockSetScript.Run(ctx, impl.redisCli, []string{impl.accountRedisKey(account)}, key, totalKey, coins, flag).Err()
}

func (impl *redisLockerImpl) Get(ctx context.Context, account, key string) (coins int64, exists bool, err error) {
	coins, err = impl.redisCli.HGet(ctx, impl.accountRedisKey(account), key).Int64()
	if err == nil {
		exists = true

		return
	}

	if errors.Is(err, redis.Nil) {
		err = nil
	}

	return
}

func (impl *redisLockerImpl) Rem(ctx context.Context, account, key string) error {
	return lockRemoveScript.Run(ctx, impl.redisCli, []string{impl.accountRedisKey(account)}, key, totalKey).Err()
}

func (impl *redisLockerImpl) GetTotal(ctx context.Context, account string) (int64, error) {
	total, err := impl.redisCli.HGet(ctx, impl.accountRedisKey(account), totalKey).Int64()
	if errors.Is(err, redis.Nil) {
		err = nil
	}

	return total, err
}

func (impl *redisLockerImpl) TransToLocker(ctx context.Context, fromAccount, fromKey string, toLocker Locker, toAccount, toKey string, options ...Option) error {
	redisToLocker, ok := toLocker.(*redisLockerImpl)
	if !ok {
		return ErrInvalidObject
	}

	flag, err := optionNew(options...).ConflictFlag()
	if err != nil {
		return err
	}

	return lockTransferScript.Run(ctx, impl.redisCli, []string{impl.accountRedisKey(fromAccount), redisToLocker.accountRedisKey(toAccount)},
		fromKey, totalKey, toKey, totalKey, flag).Err()
}

func (impl *redisLockerImpl) TransToWallet(ctx context.Context, account, key string, wallet Wallet, walletAccount, remark string) error {
	redisWallet, ok := wallet.(*redisWalletImpl)
	if !ok {
		return ErrInvalidObject
	}

	redisHistory, ok := wallet.GetHistory().(*redisHistoryImpl)
	if !ok {
		return ErrInvalidObject
	}

	return lockerTrans2WalletScript.Run(ctx, impl.redisCli, []string{impl.accountRedisKey(account), redisWallet.walletRedisKey(),
		redisHistory.accountRedisKey(account)}, key, totalKey, walletAccount, time.Now().Unix(), BuildHistoryValuePayload(account, key, remark)).Err()
}
