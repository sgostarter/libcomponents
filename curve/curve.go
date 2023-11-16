package curve

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/sgostarter/i/l"
	"github.com/sgostarter/libeasygo/routineman"
	"github.com/sgostarter/libeasygo/timespan"
	"github.com/spf13/cast"
)

type Curve[D, POINT any] struct {
	logger l.Wrapper

	baseDurationUnit time.Duration
	speeds           []int
	maxPointCount    int
	storage          Storage[POINT]
	bizSystem        BizSystem[D, POINT]

	speedTimeSpans map[int]*timespan.TimeSpan

	routineMan routineman.RoutineMan

	dHistoryLock sync.RWMutex
	dHistory     map[string][]*PointWithTimestamp[POINT]

	cachedDs *cache.Cache
}

func NewCurve[D, POINT any](baseDurationUnit time.Duration, maxPointCount int, speeds []int,
	storage Storage[POINT], bizSystem BizSystem[D, POINT], logger l.Wrapper) *Curve[D, POINT] {
	if logger == nil {
		logger = l.NewNopLoggerWrapper()
	}

	if baseDurationUnit <= 0 {
		baseDurationUnit = time.Minute
	}

	if len(speeds) == 0 {
		speeds = []int{1}
	}

	if storage == nil || bizSystem == nil {
		logger.Fatal("no dependency objects")
	}

	speedTimeSpans := make(map[int]*timespan.TimeSpan)

	for _, speed := range speeds {
		if speed <= 0 {
			logger.Fatal("invalid speed:", speed)
		}

		speedTimeSpans[speed] = timespan.NewTimeSpan(baseDurationUnit * time.Duration(speed))
	}

	cacheDuration := baseDurationUnit * 2
	if cacheDuration < time.Second {
		cacheDuration = time.Second
	}

	c := &Curve[D, POINT]{
		logger:           logger,
		baseDurationUnit: baseDurationUnit,
		speeds:           speeds,
		maxPointCount:    maxPointCount,
		storage:          storage,
		bizSystem:        bizSystem,
		speedTimeSpans:   speedTimeSpans,
		routineMan:       routineman.NewRoutineMan(context.Background(), logger),
		dHistory:         make(map[string][]*PointWithTimestamp[POINT]),
		cachedDs:         cache.New(cacheDuration, cacheDuration),
	}

	c.init()

	return c
}

func (impl *Curve[D, POINT]) TriggerStop() {
	impl.routineMan.TriggerStop()
}

func (impl *Curve[D, POINT]) Wait() {
	impl.routineMan.Wait()
}

func (impl *Curve[D, POINT]) init() {
	for _, key := range impl.bizSystem.GetKeys() {
		for _, speed := range impl.speeds {
			speedKey := impl.genStorageKey(speed, key)

			dHistory, err := impl.storage.Load(speedKey)
			if err != nil {
				continue
			}

			impl.dHistory[speedKey] = dHistory
		}
	}

	impl.routineMan.StartRoutine(impl.statisticRoutine, "statisticRoutine")
}

func (impl *Curve[D, POINT]) genStorageKey(speed int, key string) string {
	if speed == 1 {
		return key
	}

	return fmt.Sprintf("%d-%s", speed, key)
}

func (impl *Curve[D, POINT]) genCachedKey(speed int, timeLabel string) string {
	return fmt.Sprintf("%d:%s", speed, timeLabel)
}

func (impl *Curve[D, POINT]) getCachedDsOnCurrent(speed int, t time.Time) *sync.Map {
	s := impl.genCachedKey(speed, impl.speedTimeSpans[speed].GetLabel(t))

	if i, ok := impl.cachedDs.Get(s); ok {
		m, _ := i.(*sync.Map)

		return m
	}

	m := sync.Map{}
	impl.cachedDs.Set(s, &m, impl.baseDurationUnit*time.Duration(speed)*2)

	return &m
}

func (impl *Curve[D, POINT]) SetData(k string, d D) {
	t := time.Now()

	for _, speed := range impl.speeds {
		m := impl.getCachedDsOnCurrent(speed, t)

		if v, ok := m.Load(k); ok {
			// nolint:forcetypeassert
			m.Store(k, v.(ImmutableData[D]).Combine(d))
		} else {
			m.Store(k, impl.bizSystem.NewImmutableData(d))
		}
	}
}

func (impl *Curve[D, POINT]) statisticRoutine(ctx context.Context, _ func() bool) {
	speedLabels := make(map[int]string)

	for _, speed := range impl.speeds {
		speedLabels[speed] = impl.speedTimeSpans[speed].GetCurrentLabel()
	}

	sleepDuration := time.Second * 10
	if impl.baseDurationUnit/2 < sleepDuration {
		sleepDuration = impl.baseDurationUnit / 2
	}

	loop := true

	for loop {
		select {
		case <-ctx.Done():
			loop = false

			continue
		case <-time.After(sleepDuration):
			for _, speed := range impl.speeds {
				oldLabel := speedLabels[speed]
				newLabel := impl.speedTimeSpans[speed].GetCurrentLabel()

				if oldLabel == newLabel {
					continue
				}

				speedLabels[speed] = newLabel

				var currentDataSet map[string]POINT

				i, ok := impl.cachedDs.Get(impl.genCachedKey(speed, oldLabel))
				if !ok {
					continue
				}

				m, ok := i.(*sync.Map)
				if !ok {
					impl.logger.Fatal("logic error: not a map")

					continue
				}

				mm := make(map[string]D)

				m.Range(func(key, value any) bool {
					// nolint:forcetypeassert
					mm[cast.ToString(key)] = value.(ImmutableData[D]).Calc() // the value is immutable

					return true
				})

				currentDataSet = impl.bizSystem.ExplainDataAt(mm)

				t, _ := impl.speedTimeSpans[speed].Label2Time(oldLabel)

				for key, point := range currentDataSet {
					storageKey := impl.genStorageKey(speed, key)

					impl.dHistoryLock.Lock()

					impl.dHistory[storageKey] = append(impl.dHistory[storageKey], &PointWithTimestamp[POINT]{
						At: t.Unix(),
						D:  point,
					})

					if len(impl.dHistory[storageKey]) >= impl.maxPointCount {
						impl.dHistory[storageKey] = append([]*PointWithTimestamp[POINT]{}, impl.dHistory[storageKey][1:]...)
					}

					impl.dHistoryLock.Unlock()

					impl.dHistoryLock.RLock()

					_ = impl.storage.Save(storageKey, impl.dHistory[storageKey])

					impl.dHistoryLock.RUnlock()
				}
			}
		}
	}
}

func (impl *Curve[D, POINT]) GetCurves(speed int, k string, pointCnt int) (tss []int64, ps []POINT) {
	timeSpan := impl.speedTimeSpans[speed]
	if timeSpan == nil {
		return
	}

	ts, _ := timeSpan.Label2Time(timeSpan.GetCurrentLabel())

	var points []*PointWithTimestamp[POINT]

	storageKey := impl.genStorageKey(speed, k)

	impl.dHistoryLock.RLock()

	for idx := len(impl.dHistory[storageKey]) - 1; idx >= 0; idx-- {
		points = append(points, impl.dHistory[storageKey][idx])
	}

	impl.dHistoryLock.RUnlock()

	var idx int

	for ts = ts.Add(-impl.baseDurationUnit * time.Duration(speed)); pointCnt > 0; pointCnt-- {
		tss = append(tss, ts.Unix())

		if idx >= len(points)-1 || ts.Unix() != points[idx].At {
			ps = append(ps, impl.bizSystem.NewPOINT())
		} else {
			ps = append(ps, points[idx].D)
			idx++
		}

		ts = ts.Add(-impl.baseDurationUnit * time.Duration(speed))
	}

	if len(ps) == 0 {
		return
	}

	if impl.bizSystem.IsEmptyPOINT(ps[0]) {
		tss = tss[1:]
		ps = ps[1:]
	}

	for i, j := 0, len(tss)-1; i < j; i, j = i+1, j-1 {
		tss[i], tss[j] = tss[j], tss[i]
		ps[i], ps[j] = ps[j], ps[i]
	}

	return
}
