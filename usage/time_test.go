package usage

import (
	"context"
	"math/rand"
	"testing"
	"time"

	"github.com/spf13/cast"
)

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
		case <-time.After(time.Millisecond * 10):
			// nolint: gosec
			status := statuses[rand.Int31n(4)]
			ts.Update(status, 1)

			statusMap[curStatus] += time.Since(lastAt)
			lastAt = time.Now()
			curStatus = status
		}
	}

	end := time.Now()
	statusMap[curStatus] += time.Since(lastAt)

	du := ts.GetStatusStatisticsAndCleanEx(start, end, statuses, func(dOld, dNew interface{}) interface{} {
		return cast.ToInt(dOld) + cast.ToInt(dNew)
	})

	for idx := 0; idx < len(statuses); idx++ {
		t.Logf("%d: %v VS %v == %d\n", statuses[idx], du[idx].Duration, statusMap[statuses[idx]], cast.ToInt(du[idx].D))
	}
}
