package wallet

import (
	"strconv"
	"strings"
	"time"
)

func BuildHistoryValuePayload(account, key, remark string) string {
	return strconv.FormatInt(time.Now().Unix(), 10) + "\n" + account + "\n" + key + "\n" + remark
}

func ParseHistoryItem(s string) (coins int64, account, key, remark string, err error) {
	ps := strings.Split(s, "\n")
	if len(ps) != 4 {
		err = ErrBadData

		return
	}

	coins, err = strconv.ParseInt(ps[0], 10, 64)
	if err != nil {
		return
	}

	account = ps[1]
	key = ps[2]
	remark = ps[3]

	return
}
