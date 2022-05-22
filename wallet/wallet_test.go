package wallet

import (
	"context"
	"fmt"
	"testing"
	"time"

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

// nolint
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

type utHistoryStorage struct {
	items []string
	cnt   int
}

func (stg *utHistoryStorage) Store(at time.Time, item string) (err error) {
	if stg.cnt <= 0 {
		return ErrStop
	}

	stg.items = append(stg.items, item)

	stg.cnt--

	return nil
}

// nolint
func TestRedisWallet4(t *testing.T) {
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

	for idx := int64(1); idx <= 10; idx++ {
		err = wallet1.TransToWallet(context.Background(), user1, idx, fmt.Sprintf("1to2 %d", idx),
			wallet1, user2, fmt.Sprintf("2from1 %d", idx), AllowNegativeOption())
		assert.Nil(t, err)
	}

	fnCheckItems := func(items []*HistoryItem, expectCount int, coinFirst, coinLast int64, remarkFirst, remarkLast string) {
		assert.EqualValues(t, expectCount, len(items))
		if expectCount < 1 {
			return
		}

		assert.EqualValues(t, coinFirst, items[0].Coins)
		assert.EqualValues(t, remarkFirst, items[0].Remark)

		assert.EqualValues(t, coinLast, items[len(items)-1].Coins)
		assert.EqualValues(t, remarkLast, items[len(items)-1].Remark)
	}

	items, err := wallet1.GetHistory().GetItems(context.Background(), user2, 0, 1000)
	assert.Nil(t, err)
	fnCheckItems(items, 10, 10, 1, "2from1 10", "2from1 1")

	stg := &utHistoryStorage{
		cnt: 0,
	}

	err = wallet1.GetHistory().Trans2CodeStorage(user2, stg)
	assert.Nil(t, err)
	assert.EqualValues(t, 0, len(stg.items))

	items, err = wallet1.GetHistory().GetItems(context.Background(), user2, 0, 1000)
	assert.Nil(t, err)
	fnCheckItems(items, 10, 10, 1, "2from1 10", "2from1 1")

	stg = &utHistoryStorage{
		cnt: 1,
	}

	err = wallet1.GetHistory().Trans2CodeStorage(user2, stg)
	assert.Nil(t, err)
	assert.EqualValues(t, 1, len(stg.items))

	items, err = wallet1.GetHistory().GetItems(context.Background(), user2, 0, 1000)
	assert.Nil(t, err)
	fnCheckItems(items, 9, 10, 2, "2from1 10", "2from1 2")

	stg = &utHistoryStorage{
		cnt: 2,
	}

	err = wallet1.GetHistory().Trans2CodeStorage(user2, stg)
	assert.Nil(t, err)
	assert.EqualValues(t, 2, len(stg.items))

	items, err = wallet1.GetHistory().GetItems(context.Background(), user2, 0, 1000)
	assert.Nil(t, err)
	fnCheckItems(items, 7, 10, 4, "2from1 10", "2from1 4")

	stg = &utHistoryStorage{
		cnt: 3,
	}

	err = wallet1.GetHistory().Trans2CodeStorage(user2, stg)
	assert.Nil(t, err)
	assert.EqualValues(t, 3, len(stg.items))

	items, err = wallet1.GetHistory().GetItems(context.Background(), user2, 0, 1000)
	assert.Nil(t, err)
	fnCheckItems(items, 4, 10, 7, "2from1 10", "2from1 7")

	stg = &utHistoryStorage{
		cnt: 9999,
	}

	err = wallet1.GetHistory().Trans2CodeStorage(user2, stg)
	assert.Nil(t, err)
	assert.EqualValues(t, 4, len(stg.items))

	items, err = wallet1.GetHistory().GetItems(context.Background(), user2, 0, 1000)
	assert.Nil(t, err)
	fnCheckItems(items, 0, 0, 0, "", "")
}
