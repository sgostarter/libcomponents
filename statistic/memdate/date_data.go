package memdate

import (
	"time"

	"github.com/jinzhu/now"
)

type WeekD[TotalT any] struct {
	TotalT *TotalT `json:"totalT"`
	Day    int     `json:"day"`
}

type WeekData[TotalT any] struct {
	TotalD *TotalT                `json:"totalD,omitempty"`
	Day    map[int]*WeekD[TotalT] `json:"day,omitempty"`
}

func NewWeekData[TotalT any]() *WeekData[TotalT] {
	var d TotalT

	return &WeekData[TotalT]{
		TotalD: &d,
		Day:    make(map[int]*WeekD[TotalT]),
	}
}

type MonthData[TotalT any] struct {
	TotalD *TotalT                   `json:"totalD"`
	Week   map[int]*WeekData[TotalT] `json:"week"`
}

func NewMonthData[TotalT any](year, month int, loc *time.Location) *MonthData[TotalT] {
	var d TotalT

	md := &MonthData[TotalT]{
		TotalD: &d,
		Week:   make(map[int]*WeekData[TotalT]),
	}

	cT := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, loc)

	dS := make(map[int]map[int]*WeekD[TotalT]) // week,day[1-7]

	week := 1

	for ; int(cT.Month()) == month; cT = cT.Add(time.Hour * 24) {
		weekDay := cT.Weekday()

		if _, ok := dS[week]; !ok {
			dS[week] = make(map[int]*WeekD[TotalT])
		}

		var dd WeekD[TotalT]
		dd.Day = cT.Day()

		var td TotalT
		dd.TotalT = &td

		dS[week][int(weekDay)] = &dd

		if weekDay == time.Sunday {
			week++
		}
	}

	for w, weekD := range dS {
		if _, ok := md.Week[w]; !ok {
			md.Week[w] = NewWeekData[TotalT]()
		}

		for weekDay, dd := range weekD {
			md.Week[w].Day[weekDay] = dd
		}
	}

	return md
}

type SeasonData[TotalT any] struct {
	TotalD *TotalT                    `json:"totalD"`
	Month  map[int]*MonthData[TotalT] `json:"month,omitempty"`
}

func NewSeasonData[TotalT any](year, season int, loc *time.Location) *SeasonData[TotalT] {
	var d TotalT

	monthStart := (season-1)*3 + 1

	return &SeasonData[TotalT]{
		TotalD: &d,
		Month: map[int]*MonthData[TotalT]{
			monthStart:     NewMonthData[TotalT](year, monthStart, loc),
			monthStart + 1: NewMonthData[TotalT](year, monthStart+1, loc),
			monthStart + 2: NewMonthData[TotalT](year, monthStart+2, loc),
		},
	}
}

type YearData[TotalT any] struct {
	TotalD *TotalT                     `json:"totalD,omitempty"`
	Season map[int]*SeasonData[TotalT] `json:"season,omitempty"`
}

func NewYearData[TotalT any](year int, loc *time.Location) *YearData[TotalT] {
	var d TotalT

	return &YearData[TotalT]{
		TotalD: &d,
		Season: map[int]*SeasonData[TotalT]{
			1: NewSeasonData[TotalT](year, 1, loc),
			2: NewSeasonData[TotalT](year, 2, loc),
			3: NewSeasonData[TotalT](year, 3, loc),
			4: NewSeasonData[TotalT](year, 4, loc),
		},
	}
}

func GetKeysForAt(t time.Time) (year, season, month, week, weekDay int, ok bool) {
	year = t.Year()
	season = int(now.With(t).Quarter())
	month = int(t.Month())

	cT := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, t.Location())

	week, weekDay = 1, int(time.Monday)

	for ; int(cT.Month()) == month; cT = cT.Add(time.Hour * 24) {
		weekDay = int(cT.Weekday())

		if cT.Day() == t.Day() {
			ok = true

			break
		}

		if weekDay == int(time.Sunday) {
			week++
		}
	}

	return
}
