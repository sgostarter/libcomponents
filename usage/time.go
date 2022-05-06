package usage

import (
	"sync"
	"time"
)

type MergeData func(dOld, dNew interface{}) interface{}

func MergeReplace(_, dNew interface{}) interface{} {
	return dNew
}

func MergeIgnore(dOld, _ interface{}) interface{} {
	return dOld
}

type StatusTimeUsageData struct {
	Duration time.Duration
	D        interface{}
}

type TimeUsage interface {
	Update(status int, d interface{})

	DoStatisticsAndClean(tsB, tsE time.Time) map[int]*StatusTimeUsageData
	DoStatisticsAndCleanEx(tsB, tsE time.Time, md MergeData) map[int]*StatusTimeUsageData

	GetStatusStatisticsAndClean(tsB, tsE time.Time, statuses []int) []*StatusTimeUsageData
	GetStatusStatisticsAndCleanEx(tsB, tsE time.Time, statuses []int, md MergeData) []*StatusTimeUsageData
}

func NewTimeUsage() TimeUsage {
	return &timeUsageImpl{}
}

type timeUsageImpl struct {
	sync.Mutex

	ds []statusWithTime

	startStatus int
	startD      interface{}
}

type statusWithTime struct {
	status int
	d      interface{}
	at     time.Time
}

func (ts *timeUsageImpl) Update(status int, d interface{}) {
	ts.Lock()
	defer ts.Unlock()

	ts.ds = append(ts.ds, statusWithTime{
		status: status,
		d:      d,
		at:     time.Now(),
	})
}

func (ts *timeUsageImpl) DoStatistics(tsB, tsE time.Time, clearData bool, md MergeData) (ds map[int]*StatusTimeUsageData) {
	if md == nil {
		md = MergeReplace
	}

	ts.Lock()
	defer ts.Unlock()

	ds = make(map[int]*StatusTimeUsageData)

	fnD := func(status int, data interface{}) *StatusTimeUsageData {
		if d, ok := ds[status]; ok {
			d.D = md(d.D, data)

			return d
		}

		d := &StatusTimeUsageData{
			D: []interface{}{data},
		}
		ds[status] = d

		return d
	}

	if len(ts.ds) == 0 {
		fnD(ts.startStatus, ts.startD).Duration = tsE.Sub(tsB)
	} else {
		last := tsB
		for _, f := range ts.ds {
			fnD(ts.startStatus, ts.startD).Duration += f.at.Sub(last)
			last = f.at
			ts.startStatus = f.status
			ts.startD = f.d
		}
		fnD(ts.startStatus, ts.startD).Duration += tsE.Sub(last)
	}

	if clearData {
		ts.ds = ts.ds[0:0]
	}

	return
}

func (ts *timeUsageImpl) DoStatisticsAndClean(tsB, tsE time.Time) map[int]*StatusTimeUsageData {
	return ts.DoStatisticsAndCleanEx(tsB, tsE, nil)
}

func (ts *timeUsageImpl) DoStatisticsAndCleanEx(tsB, tsE time.Time, md MergeData) map[int]*StatusTimeUsageData {
	return ts.DoStatistics(tsB, tsE, true, md)
}

func (ts *timeUsageImpl) GetStatusStatisticsEx(tsB, tsE time.Time, statuses []int, clearData bool, md MergeData) []*StatusTimeUsageData {
	ds := ts.DoStatistics(tsB, tsE, clearData, md)

	statusMap := make(map[int]*StatusTimeUsageData)
	for s, d := range ds {
		statusMap[s] = d
	}

	vs := make([]*StatusTimeUsageData, len(statuses))
	for idx := 0; idx < len(statuses); idx++ {
		vs[idx] = statusMap[statuses[idx]]
	}

	return vs
}

func (ts *timeUsageImpl) GetStatusStatisticsAndClean(tsB, tsE time.Time, statuses []int) []*StatusTimeUsageData {
	return ts.GetStatusStatisticsAndCleanEx(tsB, tsE, statuses, nil)
}

func (ts *timeUsageImpl) GetStatusStatisticsAndCleanEx(tsB, tsE time.Time, statuses []int, md MergeData) []*StatusTimeUsageData {
	return ts.GetStatusStatisticsEx(tsB, tsE, statuses, true, md)
}
