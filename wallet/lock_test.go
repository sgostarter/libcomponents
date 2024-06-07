// nolint
package wallet

import (
	"context"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/sgostarter/libconfig/ut"
	"github.com/stretchr/testify/assert"
)

func initRedis(dsn string) (cli *redis.Client, err error) {
	options, err := redis.ParseURL(dsn)
	if err != nil {
		return
	}

	cli = redis.NewClient(options)

	ctx, cf := context.WithTimeout(context.Background(), 3*time.Second)
	defer cf()

	err = cli.Ping(ctx).Err()
	if err != nil {
		return
	}

	return
}

// nolint: funlen
func TestLockSet(t *testing.T) {
	cfg := ut.SetupUTConfig4Redis(t)
	redisCli, err := initRedis(cfg.RedisDSN)
	assert.Nil(t, err)

	user := "id"
	key := "test"

	//
	redisCli.Del(context.Background(), "x:locker:"+user)

	//
	locker := NewRedisLocker(redisCli, "x")

	//
	coins, exists, err := locker.Get(context.Background(), user, key)
	assert.Nil(t, err)
	assert.False(t, exists)
	assert.EqualValues(t, 0, coins)

	total, err := locker.GetTotal(context.Background(), user)
	assert.Nil(t, err)
	assert.EqualValues(t, 0, total)

	//
	err = locker.Set(context.Background(), user, key, 10)
	assert.Nil(t, err)

	coins, exists, err = locker.Get(context.Background(), user, key)
	assert.Nil(t, err)
	assert.True(t, exists)
	assert.EqualValues(t, 10, coins)

	total, err = locker.GetTotal(context.Background(), user)
	assert.Nil(t, err)
	assert.EqualValues(t, 10, total)

	//
	err = locker.Set(context.Background(), user, key, 5)
	assert.NotNil(t, err)

	coins, exists, err = locker.Get(context.Background(), user, key)
	assert.Nil(t, err)
	assert.True(t, exists)
	assert.EqualValues(t, 10, coins)

	//
	err = locker.Set(context.Background(), user, key, 6, AccumulationIfExistsOption())
	assert.Nil(t, err)

	coins, exists, err = locker.Get(context.Background(), user, key)
	assert.Nil(t, err)
	assert.True(t, exists)
	assert.EqualValues(t, 16, coins)

	total, err = locker.GetTotal(context.Background(), user)
	assert.Nil(t, err)
	assert.EqualValues(t, 16, total)

	//
	err = locker.Set(context.Background(), user, key, 6, OverflowIfExistsOption())
	assert.Nil(t, err)

	coins, exists, err = locker.Get(context.Background(), user, key)
	assert.Nil(t, err)
	assert.True(t, exists)
	assert.EqualValues(t, 6, coins)

	total, err = locker.GetTotal(context.Background(), user)
	assert.Nil(t, err)
	assert.EqualValues(t, 6, total)

	//
	err = locker.Set(context.Background(), user, key+"X", 7, AccumulationIfExistsOption())
	assert.Nil(t, err)

	coins, exists, err = locker.Get(context.Background(), user, key+"X")
	assert.Nil(t, err)
	assert.True(t, exists)
	assert.EqualValues(t, 7, coins)

	total, err = locker.GetTotal(context.Background(), user)
	assert.Nil(t, err)
	assert.EqualValues(t, 13, total)
	//
	err = locker.Rem(context.Background(), user, key)
	assert.Nil(t, err)

	total, err = locker.GetTotal(context.Background(), user)
	assert.Nil(t, err)
	assert.EqualValues(t, 7, total)
}

func TestLockTransfer(t *testing.T) {
	cfg := ut.SetupUTConfig4Redis(t)
	redisCli, err := initRedis(cfg.RedisDSN)
	assert.Nil(t, err)

	user1 := "id"
	user2 := "id2"
	//
	redisCli.Del(context.Background(), "x:locker:"+user1)
	redisCli.Del(context.Background(), "x:locker:"+user2)

	locker := NewRedisLocker(redisCli, "x")

	err = locker.Set(context.Background(), user1, "key1", 5)
	assert.Nil(t, err)
	err = locker.Set(context.Background(), user1, "key11", 6)
	assert.Nil(t, err)
	err = locker.Set(context.Background(), user1, "key111", 10)
	assert.Nil(t, err)

	//
	err = locker.TransToLocker(context.Background(), user1, "key1", locker, user2, "key2")
	assert.Nil(t, err)

	total, err := locker.GetTotal(context.Background(), user1)
	assert.Nil(t, err)
	assert.EqualValues(t, 16, total)

	total, err = locker.GetTotal(context.Background(), user2)
	assert.Nil(t, err)
	assert.EqualValues(t, 5, total)

	//
	err = locker.TransToLocker(context.Background(), user1, "key11", locker, user2, "key2")
	assert.NotNil(t, err)

	err = locker.TransToLocker(context.Background(), user1, "key11", locker, user2, "key2", AccumulationIfExistsOption())
	assert.Nil(t, err)

	total, err = locker.GetTotal(context.Background(), user1)
	assert.Nil(t, err)
	assert.EqualValues(t, 10, total)

	total, err = locker.GetTotal(context.Background(), user2)
	assert.Nil(t, err)
	assert.EqualValues(t, 11, total)

	err = locker.TransToLocker(context.Background(), user1, "key111", locker, user2, "key2", OverflowIfExistsOption())
	assert.NotNil(t, err)

	total, err = locker.GetTotal(context.Background(), user1)
	assert.Nil(t, err)
	assert.EqualValues(t, 10, total)

	total, err = locker.GetTotal(context.Background(), user2)
	assert.Nil(t, err)
	assert.EqualValues(t, 11, total)
}
