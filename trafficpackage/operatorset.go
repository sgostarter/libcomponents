package trafficpackage

import (
	"sync"
	"time"
)

func NewOperatorSet(stableID string, consumeEvent FNConsumeEvent, operators ...Operator) Operator {
	return &operatorSetImpl{
		stableID:     stableID,
		consumeEvent: consumeEvent,
		operators:    operators,
	}
}

type operatorSetImpl struct {
	lock         sync.Mutex
	stableID     string
	consumeEvent FNConsumeEvent
	operators    []Operator
}

func (impl *operatorSetImpl) GetStableID() string {
	return impl.stableID
}

func (impl *operatorSetImpl) TryConsumeAmount(id uint64, now time.Time, n int64, at time.Time, note string) (int64, error) {
	impl.lock.Lock()
	defer impl.lock.Unlock()

	var rn int64

	for _, operator := range impl.operators {
		cn, err := operator.TryConsumeAmount(id, now, n-rn, at, note)
		if err != nil {
			return 0, err
		}

		if impl.consumeEvent != nil {
			impl.consumeEvent(ConsumeTryEvent{
				TryConsumeCount: n - rn,
				ConsumedCount:   cn,
				StableID:        operator.GetStableID(),
				At:              now,
			})
		}

		rn += cn

		if rn >= n {
			break
		}
	}

	return rn, nil
}
