package wallet

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type HistoryType int

const (
	HistoryTypeWW HistoryType = iota
	HistoryTypeWL
)

// COINS\nTYPE\nTIME\nME_WALLET\nHE_WALLET_OR_LOCK\nREMARK

func BuildHistoryPayload(t HistoryType, me, he, remark string) string {
	return fmt.Sprintf("%d\n%d\n%s\n%s\n%s", t, time.Now().Unix(), me, he, remark)
}

func ParseHistoryItem(s string) (t HistoryType, at time.Time, coins int64, account, key, remark string, err error) {
	ps := strings.SplitN(s, "\n", 6)
	if len(ps) != 6 {
		err = ErrBadData

		return
	}

	n, err := strconv.ParseInt(ps[0], 10, 64)
	if err != nil {
		return
	}

	coins = n

	n, err = strconv.ParseInt(ps[1], 10, 64)
	if err != nil {
		return
	}

	t = HistoryType(n)

	n, err = strconv.ParseInt(ps[2], 10, 64)
	if err != nil {
		return
	}

	at = time.Unix(n, 0)

	account = ps[3]
	key = ps[4]
	remark = ps[5]

	return
}
