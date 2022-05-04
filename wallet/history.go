package wallet

import (
	"context"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/spf13/cast"
)

var (
	NullTime = time.Time{}
)

func newRedisHistory(redisCli *redis.Client, accountPre string) *redisHistoryImpl {
	return &redisHistoryImpl{
		redisCli:   redisCli,
		accountPre: accountPre,
	}
}

type redisHistoryImpl struct {
	redisCli   *redis.Client
	accountPre string
}

func (impl *redisHistoryImpl) accountRedisKey(account string) string {
	accountRedisKey := "history:" + account
	if impl.accountPre != "" {
		accountRedisKey = impl.accountPre + ":" + accountRedisKey
	}

	return accountRedisKey
}

func (impl *redisHistoryImpl) GetItems(ctx context.Context, account string, startAt, finishAt time.Time, offset, count int64) (items []*HistoryItem, err error) {
	fnGenTime := func(t time.Time, nullTimeString string) string {
		if t == NullTime {
			return nullTimeString
		}

		return strconv.FormatInt(t.Unix(), 10)
	}

	rItems, err := impl.redisCli.ZRangeByScoreWithScores(ctx, impl.accountRedisKey(account), &redis.ZRangeBy{
		Min:    fnGenTime(startAt, "-inf"),
		Max:    fnGenTime(finishAt, "+inf"),
		Offset: offset,
		Count:  count,
	}).Result()
	if err != nil {
		return
	}

	items = make([]*HistoryItem, 0, len(rItems))

	for _, item := range rItems {
		coins, _, _, _, _ := ParseHistoryItem(cast.ToString(item.Member))

		items = append(items, &HistoryItem{
			Coins: coins,
			At:    time.Unix(int64(item.Score), 0),
		})
	}

	return
}
