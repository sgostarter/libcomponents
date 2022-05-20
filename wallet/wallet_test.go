package wallet

import (
	"context"
	"testing"

	"github.com/sgostarter/libconfig/ut"
	"github.com/stretchr/testify/assert"
)

func TestRedisWallet(t *testing.T) {
	cfg := ut.SetupUTConfig4Redis(t)
	redisCli, err := initRedis(cfg.RedisDNS)
	assert.Nil(t, err)

	user := "user"

	redisCli.Del(context.Background(), "x:locker:"+user)
	redisCli.Del(context.Background(), "8:wallet")
	redisCli.Del(context.Background(), "8:history:"+user)

	wallet := NewRedisWallet(redisCli, "8")
	locker := NewRedisLocker(redisCli, "x")

	err = locker.Set(context.Background(), user, "key1", 5)
	assert.Nil(t, err)

	err = locker.Set(context.Background(), user, "key2", 6)
	assert.Nil(t, err)

	err = locker.TransToWallet(context.Background(), user, "key1", wallet, user, "key1 to wallet")
	assert.Nil(t, err)

	total, err := locker.GetTotal(context.Background(), user)
	assert.Nil(t, err)
	assert.EqualValues(t, 6, total)

	total, err = wallet.GetCoins(context.Background(), user)
	assert.Nil(t, err)
	assert.EqualValues(t, 5, total)

	err = wallet.TransToLocker(context.Background(), user, 2, "remark1", locker, user, "keyx")
	assert.Nil(t, err)

	total, err = locker.GetTotal(context.Background(), user)
	assert.Nil(t, err)
	assert.EqualValues(t, 8, total)

	total, err = wallet.GetCoins(context.Background(), user)
	assert.Nil(t, err)
	assert.EqualValues(t, 3, total)
}

func TestRedisWallet2(t *testing.T) {
	cfg := ut.SetupUTConfig4Redis(t)
	redisCli, err := initRedis(cfg.RedisDNS)
	assert.Nil(t, err)

	user1 := "user1"
	user2 := "user2"

	for _, user := range []string{user1, user2} {
		redisCli.Del(context.Background(), "x:locker:"+user)
		redisCli.Del(context.Background(), "8:wallet")
		redisCli.Del(context.Background(), "8:history:"+user)
		redisCli.Del(context.Background(), "9:wallet")
		redisCli.Del(context.Background(), "9:history:"+user)
	}

	wallet1 := NewRedisWallet(redisCli, "8")
	wallet2 := NewRedisWallet(redisCli, "9")

	locker := NewRedisLocker(redisCli, "x")
	err = locker.Set(context.Background(), user1, "key1", 1000)
	assert.Nil(t, err)

	err = locker.TransToWallet(context.Background(), user1, "key1", wallet1, user1, "hoho")
	assert.Nil(t, err)

	err = wallet1.TransToWallet(context.Background(), user1, 10, "remarkFrom", wallet2, user2, "remarkTo")
	assert.Nil(t, err)

	err = wallet1.TransToWallet(context.Background(), user1, 1000, "remarkFrom", wallet2, user2, "remarkTo")
	assert.NotNil(t, err)

	err = wallet1.TransToWallet(context.Background(), user1, 1000, "remarkFrom", wallet2, user2, "remarkTo", OverflowIfExistsOption(),
		AllowNegativeOption())
	assert.Nil(t, err)
}
