package memdate

import (
	"encoding/json"
	"errors"
	"os"
	"time"

	"github.com/sgostarter/i/commerr"
	"github.com/sgostarter/i/stg"
	"github.com/sgostarter/libeasygo/stg/fs/rawfs"
	"github.com/sgostarter/libeasygo/stg/mwf"
)

type DataDo[TotalD, D any] interface {
	Combine(*TotalD, D) *TotalD
}

type Statistics[K comparable, TotalT, T any, DT DataDo[TotalT, T], S mwf.Serial, L mwf.Lock] struct {
	serial S
	lock   L

	fileName string
	storage  stg.FileStorage

	mYearData map[K]map[int]*YearData[TotalT]
	loc       *time.Location

	dataTrans DT
}

func NewMemDateStatistics[K comparable, TotalT, T any, DT DataDo[TotalT, T], S mwf.Serial, L mwf.Lock](
	serial S, lock L, loc *time.Location, fileName string, storage stg.FileStorage) *Statistics[K, TotalT, T, DT, S, L] {
	if storage == nil && fileName != "" {
		storage = rawfs.NewFSStorage("")
	}

	s := &Statistics[K, TotalT, T, DT, S, L]{
		serial:    serial,
		lock:      lock,
		fileName:  fileName,
		storage:   storage,
		mYearData: make(map[K]map[int]*YearData[TotalT]),
		loc:       loc,
	}

	err := s.init()
	if err != nil {
		return nil
	}

	return s
}

func (s *Statistics[K, TotalT, T, DT, S, L]) init() error {
	return s.load()
}

func (s *Statistics[K, TotalT, T, DT, S, L]) SetDayData(key K, at time.Time, d T) (ok bool) {
	year, season, month, week, weekDay, ok := GetKeysForAt(at)
	if !ok {
		return
	}

	s.lock.Lock()
	defer s.lock.Unlock()

	s.mustYear(key, year)

	s.mYearData[key][year].Season[season].Month[month].Week[week].Day[weekDay].TotalT =
		s.dataTrans.Combine(s.mYearData[key][year].Season[season].Month[month].Week[week].Day[weekDay].TotalT, d)
	s.mYearData[key][year].Season[season].Month[month].Week[week].TotalD =
		s.dataTrans.Combine(s.mYearData[key][year].Season[season].Month[month].Week[week].TotalD, d)
	s.mYearData[key][year].Season[season].Month[month].TotalD =
		s.dataTrans.Combine(s.mYearData[key][year].Season[season].Month[month].TotalD, d)
	s.mYearData[key][year].Season[season].TotalD =
		s.dataTrans.Combine(s.mYearData[key][year].Season[season].TotalD, d)
	s.mYearData[key][year].TotalD =
		s.dataTrans.Combine(s.mYearData[key][year].TotalD, d)

	_ = s.save()

	ok = true

	return
}

func (s *Statistics[K, TotalT, T, DT, S, L]) GetYearOn(key K, at time.Time) (totalD TotalT, exists bool) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	if _, ok := s.mYearData[key]; !ok {
		return
	}

	if _, ok := s.mYearData[key][at.Year()]; !ok {
		return
	}

	totalD = *s.mYearData[key][at.Year()].TotalD
	exists = true

	return
}

func (s *Statistics[K, TotalT, T, DT, S, L]) GetSeasonOn(key K, at time.Time) (totalD TotalT, exists bool) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	year, season, _, _, _, ok := GetKeysForAt(at)
	if !ok {
		return
	}

	if _, ok = s.mYearData[key]; !ok {
		return
	}

	if _, ok = s.mYearData[key][year]; !ok {
		return
	}

	totalD = *s.mYearData[key][year].Season[season].TotalD
	exists = true

	return
}

func (s *Statistics[K, TotalT, T, DT, S, L]) GetMonthOn(key K, at time.Time) (totalD TotalT, exists bool) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	year, season, month, _, _, ok := GetKeysForAt(at)
	if !ok {
		return
	}

	if _, ok = s.mYearData[key]; !ok {
		return
	}

	if _, ok = s.mYearData[key][year]; !ok {
		return
	}

	totalD = *s.mYearData[key][year].Season[season].Month[month].TotalD
	exists = true

	return
}

func (s *Statistics[K, TotalT, T, DT, S, L]) GetWeekOn(key K, at time.Time) (totalD TotalT, exists bool) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	year, season, month, week, _, ok := GetKeysForAt(at)
	if !ok {
		return
	}

	if _, ok = s.mYearData[key]; !ok {
		return
	}

	if _, ok = s.mYearData[key][year]; !ok {
		return
	}

	totalD = *s.mYearData[key][year].Season[season].Month[month].Week[week].TotalD
	exists = true

	return
}

func (s *Statistics[K, TotalT, T, DT, S, L]) GetDayOn(key K, at time.Time) (totalD TotalT, exists bool) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	year, season, month, week, weekD, ok := GetKeysForAt(at)
	if !ok {
		return
	}

	if _, ok = s.mYearData[key]; !ok {
		return
	}

	if _, ok = s.mYearData[key][year]; !ok {
		return
	}

	totalD = *s.mYearData[key][year].Season[season].Month[month].Week[week].Day[weekD].TotalT
	exists = true

	return
}

func (s *Statistics[K, TotalT, T, DT, S, L]) mustYear(key K, year int) {
	if _, ok := s.mYearData[key]; !ok {
		s.mYearData[key] = make(map[int]*YearData[TotalT])
	}

	if _, ok := s.mYearData[key][year]; !ok {
		s.mYearData[key][year] = NewYearData[TotalT](year, s.loc)
	}
}

func (s *Statistics[K, TotalT, T, DT, S, L]) load() error {
	if s.fileName == "" {
		return nil
	}

	d, err := s.storage.ReadFile(s.fileName)
	if err != nil {
		var pathError *os.PathError

		if errors.As(err, &pathError) {
			err = nil
		}

		return err
	}

	var m map[K]map[int]*YearData[TotalT]

	err = s.serial.Unmarshal(d, &m)
	if err != nil {
		return err
	}

	if len(m) == 0 {
		m = make(map[K]map[int]*YearData[TotalT])
	}

	s.mYearData = m

	return nil
}

func (s *Statistics[K, TotalT, T, DT, S, L]) save() error {
	if s.fileName == "" {
		return nil
	}

	d, err := s.serial.Marshal(s.mYearData)
	if err != nil {
		return err
	}

	err = s.storage.WriteFile(s.fileName, d)
	if err != nil {
		return err
	}

	return nil
}

func (s *Statistics[K, TotalT, T, DT, S, L]) Export(key K) (dM map[int]*YearData[TotalT], err error) {
	var d []byte

	s.lock.RLock()
	if yD, ok := s.mYearData[key]; ok {
		d, err = json.Marshal(yD)
	} else {
		err = commerr.ErrNotFound
	}
	s.lock.RUnlock()

	if err != nil {
		return
	}

	err = json.Unmarshal(d, &dM)

	return
}
