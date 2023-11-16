package usage

import (
	"context"
	"math/rand"
	"testing"
	"time"

	"github.com/spf13/cast"
)

// nolint
func TestTimeUsage(t *testing.T) {
	statuses := []int{1, 2, 3, 4}
	statusMap := make(map[int]time.Duration)

	ts := NewTimeUsage()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	curStatus := 0
	lastAt := time.Now()

	loop := true

	start := time.Now()

	for loop {
		select {
		case <-ctx.Done():
			loop = false

			continue
		case <-time.After(time.Duration(rand.Int63n(int64(time.Millisecond * 10)))):
			status := statuses[rand.Int31n(4)]
			ts.Update(status, 1)

			statusMap[curStatus] += time.Since(lastAt)
			lastAt = time.Now()
			curStatus = status
		}
	}

	end := time.Now()
	statusMap[curStatus] += time.Since(lastAt)

	du := ts.GetStatusStatisticsEx(start, end, statuses, false, func(dOld, dNew interface{}) interface{} {
		return cast.ToInt(dOld) + cast.ToInt(dNew)
	})

	for idx := 0; idx < len(statuses); idx++ {
		t.Logf("%d: %v VS %v == %d\n", statuses[idx], du[idx].Duration, statusMap[statuses[idx]], cast.ToInt(du[idx].D))
	}

	for idx := 0; idx < 100; idx++ {
		status := statuses[rand.Int31n(4)]

		time.Sleep(time.Millisecond * 10)
		ts.Update(status, 1)
	}

	du = ts.GetStatusStatisticsEx(end, time.Now(), statuses, false, func(dOld, dNew interface{}) interface{} {
		return cast.ToInt(dOld) + cast.ToInt(dNew)
	})

	for idx := 0; idx < len(statuses); idx++ {
		t.Logf("%d: %v VS %v == %d\n", statuses[idx], du[idx].Duration, statusMap[statuses[idx]], cast.ToInt(du[idx].D))
	}

	du = ts.GetStatusStatisticsEx(start, time.Now(), statuses, false, func(dOld, dNew interface{}) interface{} {
		return cast.ToInt(dOld) + cast.ToInt(dNew)
	})

	for idx := 0; idx < len(statuses); idx++ {
		t.Logf("%d: %v VS %v == %d\n", statuses[idx], du[idx].Duration, statusMap[statuses[idx]], cast.ToInt(du[idx].D))
	}

	du = ts.GetStatusStatisticsAndCleanEx(start, end, statuses, func(dOld, dNew interface{}) interface{} {
		return cast.ToInt(dOld) + cast.ToInt(dNew)
	})

	for idx := 0; idx < len(statuses); idx++ {
		t.Logf("%d: %v VS %v == %d\n", statuses[idx], du[idx].Duration, statusMap[statuses[idx]], cast.ToInt(du[idx].D))
	}
}

func TestTimeUsage2(t *testing.T) {
	timeNow := time.Now()
	ts := NewTimeUsage()

	time.Sleep(time.Second * 2)
	ts.Update(1, nil)
	vs := ts.GetStatusStatisticsAndClean(timeNow.Add(-time.Second), timeNow.Add(time.Second), []int{1, 2})
	t.Log(vs)
}
