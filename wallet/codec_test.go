package wallet

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCodec(t *testing.T) {
	s := BuildHistoryPayload(HistoryTypeWW, "me", "he", "a")
	s = "100\n" + s
	ht, at, coins, account, key, remark, err := ParseHistoryItem(s)
	assert.Nil(t, err)
	assert.EqualValues(t, HistoryTypeWW, ht)
	assert.EqualValues(t, 100, coins)
	assert.EqualValues(t, "me", account)
	assert.EqualValues(t, "he", key)
	assert.EqualValues(t, "a", remark)
	assert.True(t, time.Since(at) < time.Second)
	assert.True(t, time.Since(at) > -time.Second)

	s = BuildHistoryPayload(HistoryTypeWW, "me", "he", "")
	s = "200\n" + s
	ht, at, coins, account, key, remark, err = ParseHistoryItem(s)
	assert.Nil(t, err)
	assert.EqualValues(t, HistoryTypeWW, ht)
	assert.EqualValues(t, 200, coins)
	assert.EqualValues(t, "me", account)
	assert.EqualValues(t, "he", key)
	assert.EqualValues(t, "", remark)
	assert.True(t, time.Since(at) < time.Second)
	assert.True(t, time.Since(at) > -time.Second)
}
