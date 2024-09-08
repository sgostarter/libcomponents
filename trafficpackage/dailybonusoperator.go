package trafficpackage

import (
	"time"

	"github.com/sgostarter/i/commerr"
	"github.com/sgostarter/i/l"
)

func NewDailyBonusOperator(storage DailyBonusStorage, logger l.Wrapper) DailyBonusOperator {
	if logger == nil {
		logger = l.NewNopLoggerWrapper()
	}

	logger = logger.WithFields(l.StringField(l.ClsKey, "dailyBonusOperatorImpl"))

	if storage == nil {
		logger.Fatal("no storage")
	}

	return &dailyBonusOperatorImpl{
		logger:  logger,
		storage: storage,
	}
}

type dailyBonusOperatorImpl struct {
	logger  l.Wrapper
	storage DailyBonusStorage
}

func (impl *dailyBonusOperatorImpl) ConsumeAmount(id uint64, now time.Time, n int64, at time.Time, note string) (err error) {
	if n <= 0 {
		return
	}

	bonus, todayBonus, err := impl.storage.GetAllBonus(id, now)
	if err != nil {
		return
	}

	var bonusValue, todayBonusValue int64

	if n > todayBonus+bonus {
		err = commerr.ErrResourceExhausted

		return
	}

	if n <= todayBonus {
		todayBonusValue = n
	} else {
		todayBonusValue = todayBonus
		n -= todayBonus
		bonusValue -= n
	}

	err = impl.storage.ConsumeBonus(id, bonusValue, todayBonusValue, at, note)

	return
}

func (impl *dailyBonusOperatorImpl) TryConsumeAmount(id uint64, now time.Time, n int64, at time.Time, note string) (rn int64, err error) {
	if n <= 0 {
		return
	}

	rn = n

	bonus, todayBonus, err := impl.storage.GetAllBonus(id, now)
	if err != nil {
		return
	}

	if bonus+todayBonus <= 0 {
		rn = 0

		return
	}

	var bonusValue, todayBonusValue int64

	if n <= todayBonus {
		todayBonusValue = n
		n = 0
	} else {
		todayBonusValue = todayBonus
		n -= todayBonus

		bonusValue = n
		if bonusValue > bonus {
			bonusValue = bonus
			n -= bonus
		} else {
			n = 0
		}
	}

	err = impl.storage.ConsumeBonus(id, bonusValue, todayBonusValue, at, note)
	if err != nil {
		return
	}

	rn -= n

	return
}

func (impl *dailyBonusOperatorImpl) HasDailyBonus(id uint64, now time.Time) (bool, error) {
	return impl.storage.HasDailyBonus(id, now)
}

func (impl *dailyBonusOperatorImpl) EarnDailyBonus(id uint64, now time.Time) error {
	return impl.storage.EarnDailyBonus(id, now)
}

func (impl *dailyBonusOperatorImpl) Get(id uint64, now time.Time) (bonus, todayBonus int64, err error) {
	return impl.storage.GetAllBonus(id, now)
}
