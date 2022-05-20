package wallet

import (
	"context"
	"time"
)

type HistoryItem struct {
	Coins int64
	At    time.Time
}

type History interface {
	GetItems(ctx context.Context, account string, startAt, finishAt time.Time, offset, count int64) ([]*HistoryItem, error)
}

type Wallet interface {
	GetHistory() History

	TransToLocker(ctx context.Context, account string, coins int64, remark string, locker Locker, toAccount, key string, options ...Option) (err error)
	TransToWallet(ctx context.Context, account string, coins int64, remarkFrom string, wallet Wallet, accountTo, remarkTo string, options ...Option) (err error)

	GetCoins(ctx context.Context, account string) (int64, error)
}

type Locker interface {
	Set(ctx context.Context, account, key string, coins int64, options ...Option) error
	Get(ctx context.Context, account, key string) (coins int64, exists bool, err error)
	Rem(ctx context.Context, account, key string) error

	GetTotal(ctx context.Context, account string) (int64, error)

	TransToLocker(ctx context.Context, fromAccount, fromKey string, toLocker Locker, toAccount, toKey string, options ...Option) error
	TransToWallet(ctx context.Context, account, key string, wallet Wallet, walletAccount string, remark string) error
}
