// nolint
package ut

import (
	"os"
	"testing"
	"time"

	"github.com/sgostarter/libcomponents/trafficpackage"
	"github.com/sgostarter/libcomponents/trafficpackage/impl/fmstorage"
	"github.com/stretchr/testify/assert"
)

func Test(t *testing.T) {
	_ = os.RemoveAll("./ut-data")

	tp := trafficpackage.NewTrafficPackage(fmstorage.NewFMStorage("ut-data", nil), nil)

	uid := uint64(10)

	amount, err := tp.GetAmount(uid)
	assert.Nil(t, err)
	assert.EqualValues(t, 0, amount)

	packageID, err := tp.AddPackage(uid, 10, time.Now())
	assert.Nil(t, err)
	assert.True(t, packageID > 0)

	amount, err = tp.GetAmount(uid)
	assert.Nil(t, err)
	assert.EqualValues(t, 10, amount)

	packageID, err = tp.AddPackage(uid, 100, time.Now())
	assert.Nil(t, err)
	assert.True(t, packageID > 0)

	amount, err = tp.GetAmount(uid)
	assert.Nil(t, err)
	assert.EqualValues(t, 110, amount)

	err = tp.ConsumeAmount(uid, time.Now(), 1, time.Now(), "hh")
	assert.Nil(t, err)

	amount, err = tp.GetAmount(uid)
	assert.Nil(t, err)
	assert.EqualValues(t, 109, amount)

	err = tp.ConsumeAmount(uid, time.Now(), 10, time.Now(), "hh")
	assert.Nil(t, err)

	amount, err = tp.GetAmount(uid)
	assert.Nil(t, err)
	assert.EqualValues(t, 99, amount)
}

func checkBonus(t *testing.T, uid uint64, now time.Time, tp trafficpackage.DailyBonusOperator, bonus, dailyBonus int64) {
	_bonus, _todayBonus, err := tp.Get(uid, now)
	assert.Nil(t, err)

	assert.EqualValues(t, bonus, _bonus)
	assert.EqualValues(t, dailyBonus, _todayBonus)
}

func TestDailyBonus(t *testing.T) {
	_ = os.RemoveAll("./ut-data")

	tp := trafficpackage.NewDailyBonusOperator(fmstorage.NewFMDailyBonusStorage("ut-data", nil), nil)

	uid := uint64(10)

	now := time.Now()

	checkBonus(t, uid, now, tp, 0, 0)

	err := tp.EarnDailyBonus(uid, now)
	assert.Nil(t, err)

	checkBonus(t, uid, now, tp, 0, 5)

	err = tp.ConsumeAmount(uid, time.Now(), 2, time.Now(), "")
	assert.Nil(t, err)

	checkBonus(t, uid, now, tp, 0, 3)

	tomorrow := now.Add(time.Hour * 24)

	checkBonus(t, uid, tomorrow, tp, 3, 0)

	err = tp.EarnDailyBonus(uid, tomorrow)
	assert.Nil(t, err)

	checkBonus(t, uid, tomorrow, tp, 3, 5)

	err = tp.EarnDailyBonus(uid, tomorrow)
	assert.NotNil(t, err)

	n, err := tp.TryConsumeAmount(uid, tomorrow, 1, time.Now(), "")
	assert.Nil(t, err)
	assert.EqualValues(t, 1, n)

	checkBonus(t, uid, tomorrow, tp, 3, 4)

	n, err = tp.TryConsumeAmount(uid, tomorrow, 9, time.Now(), "")
	assert.Nil(t, err)
	assert.EqualValues(t, 7, n)

	checkBonus(t, uid, tomorrow, tp, 0, 0)

	n, err = tp.TryConsumeAmount(uid, tomorrow, 1, time.Now(), "")
	assert.Nil(t, err)
	assert.EqualValues(t, 0, n)
}
