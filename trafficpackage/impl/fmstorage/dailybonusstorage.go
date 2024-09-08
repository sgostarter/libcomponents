package fmstorage

import (
	"path/filepath"
	"sync"
	"time"

	"github.com/sgostarter/i/commerr"
	"github.com/sgostarter/i/stg"
	"github.com/sgostarter/libcomponents/trafficpackage"
	"github.com/sgostarter/libeasygo/stg/fs/rawfs"
	"github.com/sgostarter/libeasygo/stg/mwf"
)

type FNDate func(now time.Time) string

func NewFMDailyBonusStorage(root string, storage stg.FileStorage) trafficpackage.DailyBonusStorage {
	return NewFMDailyBonusStorageEx(root, storage, "daily-bonus.json", false, nil, nil)
}

func NewFMDailyBonusStorageEx(root string, storage stg.FileStorage, fileName string, prettySerial bool,
	fnDate FNDate, fnBonusInitForNewID trafficpackage.FNDailyBonusInitForNewID) trafficpackage.DailyBonusStorage {
	if storage == nil {
		storage = rawfs.NewFSStorage("")
	}

	if fnDate == nil {
		fnDate = date4Day
	}

	impl := &fmDailyBonusStorageImpl{
		storage: mwf.NewMemWithFile[map[uint64]*dailyData, mwf.Serial, mwf.Lock](
			make(map[uint64]*dailyData), &mwf.JSONSerial{
				MarshalIndent: prettySerial,
			}, &sync.RWMutex{}, filepath.Join(root, fileName), storage),
		fnDate:              fnDate,
		fnBonusInitForNewID: fnBonusInitForNewID,
	}

	return impl
}

type dailyData struct {
	Date       string `json:"date,omitempty" yaml:"date,omitempty,omitempty"`
	Bonus      int64  `json:"bonus,omitempty" yaml:"bonus,omitempty"`
	DailyBonus int64  `json:"dailyBonus,omitempty" yaml:"dailyBonus,omitempty"`
}

type fmDailyBonusStorageImpl struct {
	storage             *mwf.MemWithFile[map[uint64]*dailyData, mwf.Serial, mwf.Lock]
	fnDate              FNDate
	fnBonusInitForNewID trafficpackage.FNDailyBonusInitForNewID
}

func date4Day(now time.Time) string {
	return now.Format("20060102")
}

func (impl *fmDailyBonusStorageImpl) initDataForNewID(dd *dailyData) error {
	if impl.fnBonusInitForNewID == nil {
		return nil
	}

	bonus, dailyBonus, err := impl.fnBonusInitForNewID()
	if err != nil {
		return err
	}

	dd.Bonus = bonus
	dd.DailyBonus = dailyBonus

	return nil
}

func (impl *fmDailyBonusStorageImpl) GetAllBonus(id uint64, now time.Time) (bonus, todayBonus int64, err error) {
	_ = impl.storage.Change(func(oldD map[uint64]*dailyData) (map[uint64]*dailyData, error) {
		newD := oldD
		if len(newD) == 0 {
			newD = make(map[uint64]*dailyData)
		}

		dd, ok := newD[id]
		if !ok {
			dd = &dailyData{}

			err = impl.initDataForNewID(dd)
			if err != nil {
				return nil, commerr.ErrReject
			}

			newD[id] = dd
		}

		if dd.DailyBonus > 0 && dd.Date != impl.fnDate(now) {
			dd.Bonus += dd.DailyBonus
			dd.DailyBonus = 0
		}

		bonus = dd.Bonus
		todayBonus = dd.DailyBonus

		return newD, nil
	})

	return
}

func (impl *fmDailyBonusStorageImpl) ConsumeBonus(id uint64, bonusValue, todayBonusValue int64, at time.Time, note string) (err error) {
	_ = impl.storage.Change(func(oldD map[uint64]*dailyData) (map[uint64]*dailyData, error) {
		newD := oldD
		if len(newD) == 0 {
			newD = make(map[uint64]*dailyData)
		}

		dd, ok := newD[id]
		if !ok {
			dd = &dailyData{}

			err = impl.initDataForNewID(dd)
			if err != nil {
				return nil, commerr.ErrReject
			}

			newD[id] = dd
		}

		if dd.DailyBonus < todayBonusValue || dd.Bonus < bonusValue {
			err = commerr.ErrOutOfRange

			return nil, commerr.ErrOutOfRange
		}

		dd.Bonus -= bonusValue
		dd.DailyBonus -= todayBonusValue

		return newD, nil
	})

	return
}

func (impl *fmDailyBonusStorageImpl) HasDailyBonus(id uint64, now time.Time) (f bool, err error) {
	_ = impl.storage.Change(func(oldD map[uint64]*dailyData) (map[uint64]*dailyData, error) {
		newD := oldD
		if len(newD) == 0 {
			newD = make(map[uint64]*dailyData)
		}

		dd, ok := newD[id]
		if !ok {
			dd = &dailyData{}

			err = impl.initDataForNewID(dd)
			if err != nil {
				return nil, commerr.ErrReject
			}

			newD[id] = dd
		}

		if dd.DailyBonus > 0 && dd.Date != impl.fnDate(now) {
			dd.Bonus += dd.DailyBonus
			dd.DailyBonus = 0
		}

		f = dd.Date == impl.fnDate(now)

		return newD, nil
	})

	return
}

func (impl *fmDailyBonusStorageImpl) EarnDailyBonus(id uint64, now time.Time) (err error) {
	_ = impl.storage.Change(func(oldD map[uint64]*dailyData) (map[uint64]*dailyData, error) {
		newD := oldD
		if len(newD) == 0 {
			newD = make(map[uint64]*dailyData)
		}

		dd, ok := newD[id]
		if !ok {
			dd = &dailyData{}

			err = impl.initDataForNewID(dd)
			if err != nil {
				return nil, commerr.ErrReject
			}

			newD[id] = dd
		}

		nowDate := impl.fnDate(now)

		if dd.DailyBonus > 0 && dd.Date != nowDate {
			dd.Bonus += dd.DailyBonus
			dd.DailyBonus = 0
		}

		if dd.Date != nowDate {
			dd.DailyBonus = 5
			dd.Date = nowDate
		} else {
			err = commerr.ErrAlreadyExists
		}

		return newD, nil
	})

	return
}
