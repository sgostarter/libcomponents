package memdate

import (
	"testing"
	"time"

	"github.com/sgostarter/libcomponents/statistic/memdate/ex"
	"github.com/sgostarter/libeasygo/stg/mwf"
	"github.com/stretchr/testify/assert"
)

func Test1(t *testing.T) {
	stat := NewMemDateStatistics[string, ex.LifeCostTotalData, ex.LifeCostData, ex.LifeCostDataTrans, mwf.Serial, mwf.Lock](
		&mwf.JSONSerial{}, &mwf.NoLock{}, time.Local, "utStorage.txt", nil)

	const key = "zjz"

	at := time.Now()

	ok := stat.SetDayData(key, at, ex.LifeCostData{
		T:             ex.ListCostDataAdd,
		ConsumeCount:  1,
		ConsumeAmount: 100,
		EarnCount:     2,
		EarnAmount:    200,
	})
	assert.True(t, ok)

	ok = stat.SetDayData(key, time.Date(2023, 12, 1, 0, 0, 0, 0, time.Local), ex.LifeCostData{
		T:             ex.ListCostDataAdd,
		ConsumeCount:  1,
		ConsumeAmount: 22,
		EarnCount:     2,
		EarnAmount:    33,
	})
	assert.True(t, ok)

	totalD, ok := stat.GetYearOn(key, at)
	assert.True(t, ok)
	t.Log(totalD)

	totalD, ok = stat.GetWeekOn(key, at)
	assert.True(t, ok)
	t.Log(totalD)

	totalD, ok = stat.GetMonthOn(key, at)
	assert.True(t, ok)
	t.Log(totalD)
}
