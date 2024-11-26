// nolint
package redisimpls

import (
	"context"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/sgostarter/libcomponents/account"
	"github.com/stretchr/testify/assert"
)

func Test1(t *testing.T) {
	opts, err := redis.ParseURL("redis://:@127.0.0.1:6379") // redis://<user>:<password>@<host>:<port>/<db_number>
	assert.Nil(t, err)

	redisCli := redis.NewClient(opts)

	redisCli.FlushDB(context.Background())

	stg := NewRedisAccountStorage("x", redisCli, nil)

	f, err := stg.HasAccount()
	assert.Nil(t, err)
	assert.False(t, f)

	uid1, err := stg.AddAccount("user1", "hpass1")
	assert.Nil(t, err)

	f, err = stg.HasAccount()
	assert.Nil(t, err)
	assert.True(t, f)

	_, err = stg.AddAccount("user1", "hpass11")
	assert.NotNil(t, err)

	_, err = stg.AddAccountEx(uid1, "user12", "hpass11", nil)
	assert.NotNil(t, err)

	start := time.Now()

	users, err := stg.ListUsers(0, 0)
	assert.Nil(t, err)
	assert.Len(t, users, 1)

	users, err = stg.ListUsers(start.Unix(), 0)
	assert.Nil(t, err)
	assert.Len(t, users, 1)
	assert.EqualValues(t, "user1", users[0].UserName)

	users, err = stg.ListUsers(start.Unix(), start.Unix())
	assert.Nil(t, err)
	assert.Len(t, users, 1)

	users, err = stg.ListUsers(start.Unix(), start.Unix()-1)
	assert.Nil(t, err)
	assert.Empty(t, users)

	accountName, password, err := stg.GetAccount(uid1)
	assert.Nil(t, err)
	assert.EqualValues(t, "user1", accountName)
	assert.EqualValues(t, "hpass1", password)

	err = stg.SetHashedPassword(uid1, "hpass2")
	assert.Nil(t, err)

	err = stg.RenameAccountName(uid1, "user2")
	assert.Nil(t, err)

	accountName, password, err = stg.GetAccount(uid1)
	assert.Nil(t, err)
	assert.EqualValues(t, "user2", accountName)
	assert.EqualValues(t, "hpass2", password)

	advCfg := account.AdvanceConfig{
		TokenExpiresAfter: time.Hour * 9,
	}

	err = stg.SetAdvanceConfig(uid1, &advCfg)
	assert.Nil(t, err)

	advCfg2, err := stg.GetAdvanceConfig(uid1)
	assert.Nil(t, err)

	assert.EqualValues(t, advCfg.TokenExpiresAfter, advCfg2.TokenExpiresAfter)

	time.Sleep(time.Second * 2)

	uidX, err := stg.AddAccountEx(0, "userX", "hPassX", []byte{0x01, 0x02})
	assert.Nil(t, err)

	assert.NotEqualValues(t, uidX, uid1)

	uidX2, _, err := stg.FindAccount("userX")
	assert.Nil(t, err)
	assert.EqualValues(t, uidX, uidX2)

	accounts, err := stg.ListUsers(0, 0)
	assert.Nil(t, err)
	assert.Len(t, accounts, 2)
	assert.EqualValues(t, uid1, accounts[0].UserID)
	assert.EqualValues(t, uidX, accounts[1].UserID)

	data, err := stg.GetAccountData(uid1)
	assert.Nil(t, err)
	assert.EqualValues(t, []byte{}, data)

	data, err = stg.GetAccountData(uidX)
	assert.Nil(t, err)
	assert.EqualValues(t, []byte{0x01, 0x02}, data)

	uidX1, exists, err := stg.GetIDFromAccountName("userX")
	assert.Nil(t, err)
	assert.True(t, exists)
	assert.EqualValues(t, uidX, uidX1)

	var pd propertyData

	err = stg.GetPropertyData("userX", &pd)
	assert.Nil(t, err)

	err = stg.SetPropertyData("userX", propertyData{
		N: 9,
		S: "9",
	})
	assert.Nil(t, err)

	err = stg.GetPropertyData("userX", &pd)
	assert.Nil(t, err)

	assert.EqualValues(t, 9, pd.N)
	assert.EqualValues(t, "9", pd.S)

	data, err = stg.GetAccountData(uidX)
	assert.Nil(t, err)
	assert.EqualValues(t, []byte{0x01, 0x02}, data)

	err = stg.AddToken("token1", uid1, time.Now().Add(time.Minute*10))
	assert.Nil(t, err)

	err = stg.AddToken("token2", uid1, time.Now().Add(time.Minute*10))
	assert.Nil(t, err)

	err = stg.AddToken("token-x", uidX, time.Now().Add(time.Minute*10))
	assert.Nil(t, err)

	exists, err = stg.TokenExists("token1", time.Minute*10)
	assert.Nil(t, err)
	assert.True(t, exists)

	exists, err = stg.TokenExists("token2", time.Minute*10)
	assert.Nil(t, err)
	assert.True(t, exists)

	err = stg.DelToken("token2")
	assert.Nil(t, err)

	err = stg.DelToken("token1")
	assert.Nil(t, err)
}

type propertyData struct {
	N int
	S string
}
