package wallet

import (
	"context"

	"github.com/go-redis/redis/v8"
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

func (impl *redisHistoryImpl) GetItems(ctx context.Context, account string, offset, count int64) (items []*HistoryItem, err error) {
	if count == 0 {
		count = 10000
	}

	return impl.getItems(ctx, account, offset, offset+count-1, false)
}

func (impl *redisHistoryImpl) GetItemsASC(ctx context.Context, account string, offset, count int64) ([]*HistoryItem, error) {
	if count == 0 {
		count = 10000
	}

	return impl.getItems(ctx, account, -offset-count, -offset-1, true)
}

func (impl *redisHistoryImpl) getItems(ctx context.Context, account string, start, stop int64, reverseOutput bool) (items []*HistoryItem, err error) {
	rItems, err := impl.redisCli.LRange(ctx, impl.accountRedisKey(account), start, stop).Result()
	if err != nil {
		return
	}

	items = make([]*HistoryItem, 0, len(rItems))

	for _, item := range rItems {
		_, at, coins, _, _, remark, _ := ParseHistoryItem(item)

		items = append(items, &HistoryItem{
			Coins:  coins,
			At:     at,
			Remark: remark,
		})
	}

	if reverseOutput {
		for idxB := 0; idxB < len(rItems)-1; idxB++ {
			items[idxB], items[len(rItems)-1-idxB] = items[len(rItems)-1-idxB], items[idxB]
		}
	}

	return
}
