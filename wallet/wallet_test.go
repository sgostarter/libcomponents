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

// nolint
func TestRedisWallet3(t *testing.T) {
	cfg := ut.SetupUTConfig4Redis(t)
	redisCli, err := initRedis(cfg.RedisDNS)
	assert.Nil(t, err)

	user1 := "user1"
	user2 := "user2"

	for _, user := range []string{user1, user2} {
		redisCli.Del(context.Background(), "8:wallet")
		redisCli.Del(context.Background(), "8:history:"+user)
	}

	wallet1 := NewRedisWallet(redisCli, "8")

	err = wallet1.TransToWallet(context.Background(), user1, 10000, "1to2 10000", wallet1, user2, "2from1 10000", AllowNegativeOption())
	assert.Nil(t, err)

	err = wallet1.TransToWallet(context.Background(), user1, 1000, "1to2 1000", wallet1, user2, "2from1 1000", AllowNegativeOption())
	assert.Nil(t, err)

	err = wallet1.TransToWallet(context.Background(), user1, 100, "1to2 100", wallet1, user2, "2from1 100", AllowNegativeOption())
	assert.Nil(t, err)

	err = wallet1.TransToWallet(context.Background(), user1, 10, "1to2 10", wallet1, user2, "2from1 10", AllowNegativeOption())
	assert.Nil(t, err)

	err = wallet1.TransToWallet(context.Background(), user1, 1, "1to2 1", wallet1, user2, "2from1 1", AllowNegativeOption())
	assert.Nil(t, err)

	items, err := wallet1.GetHistory().GetItems(context.Background(), user1, 0, 0)
	assert.Nil(t, err)
	assert.EqualValues(t, 5, len(items))

	items, err = wallet1.GetHistory().GetItems(context.Background(), user1, 0, 1)
	assert.Nil(t, err)
	assert.EqualValues(t, 1, len(items))
	assert.EqualValues(t, -1, items[0].Coins)

	items, err = wallet1.GetHistory().GetItems(context.Background(), user1, 1, 1)
	assert.Nil(t, err)
	assert.EqualValues(t, 1, len(items))
	assert.EqualValues(t, -10, items[0].Coins)

	items, err = wallet1.GetHistory().GetItems(context.Background(), user1, 0, 2)
	assert.Nil(t, err)
	assert.EqualValues(t, 2, len(items))
	assert.EqualValues(t, -1, items[0].Coins)
	assert.EqualValues(t, -10, items[1].Coins)

	items, err = wallet1.GetHistory().GetItems(context.Background(), user1, 0, 10000)
	assert.Nil(t, err)
	assert.EqualValues(t, 5, len(items))
	assert.EqualValues(t, -1, items[0].Coins)
	assert.EqualValues(t, "1to2 1", items[0].Remark)
	assert.EqualValues(t, -10000, items[4].Coins)
	assert.EqualValues(t, "1to2 10000", items[4].Remark)

	items, err = wallet1.GetHistory().GetItemsASC(context.Background(), user2, 0, 0)
	assert.Nil(t, err)
	assert.EqualValues(t, 5, len(items))

	items, err = wallet1.GetHistory().GetItemsASC(context.Background(), user2, 0, 1)
	assert.Nil(t, err)
	assert.EqualValues(t, 1, len(items))
	assert.EqualValues(t, 10000, items[0].Coins)

	items, err = wallet1.GetHistory().GetItemsASC(context.Background(), user2, 0, 10000)
	assert.Nil(t, err)
	assert.EqualValues(t, 5, len(items))
	assert.EqualValues(t, 10000, items[0].Coins)
	assert.EqualValues(t, "2from1 10000", items[0].Remark)
	assert.EqualValues(t, 1, items[4].Coins)
	assert.EqualValues(t, "2from1 1", items[4].Remark)
}
