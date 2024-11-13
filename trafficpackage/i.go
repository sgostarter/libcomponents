package trafficpackage

import "time"

type Package struct {
	ID     uint64    `json:"id" yaml:"id"`
	Amount int64     `json:"amount,omitempty" yaml:"amount,omitempty"`
	At     time.Time `json:"at,omitempty" yaml:"at,omitempty"`
}

type PackageInfo struct {
	Package

	LeftAmount int64 `json:"left_amount,omitempty" yaml:"left_amount,omitempty"`
}

type ConsumeTryEvent struct {
	TryConsumeCount int64
	ConsumedCount   int64
	StableID        string
	At              time.Time
}

type FNConsumeEvent func(e ConsumeTryEvent)

type Operator interface {
	GetStableID() string

	TryConsumeAmount(id uint64, now time.Time, n int64, at time.Time, note string) (int64, error)
}

type TrafficPackage interface {
	Operator
	ConsumeAmount(id uint64, now time.Time, n int64, at time.Time, note string) error

	GetAmount(id uint64) (int64, error)
	GetPackageInfo(id, packageID uint64) (PackageInfo, error)
	GetPackages(id uint64, includeNoDataPackages bool) ([]PackageInfo, error)

	AddPackage(id uint64, amount int64, at time.Time) (newPackageID uint64, err error)
	AddPackageEx(id, packageID uint64, amount int64, at time.Time) (newPackageID uint64, err error)
}

type ConsumeData struct {
	PackageID     uint64
	ConsumeAmount int64
	OldAmount     int64
}

type Storage interface {
	GetPackageInfo(id, packageID uint64) (PackageInfo, error)
	GetPackages(id uint64, includeNoDataPackages bool) ([]PackageInfo, error)
	Consume(id uint64, cds []ConsumeData, at time.Time, note string) error
	AddPackage(id, packageID uint64, amount int64, at time.Time) (newPackageID uint64, err error)
}

type DailyBonusOperator interface {
	Operator

	ConsumeAmount(id uint64, now time.Time, n int64, at time.Time, note string) error

	HasDailyBonus(id uint64, now time.Time) (bool, error)
	EarnDailyBonus(id uint64, now time.Time) error
	Get(id uint64, now time.Time) (bonus, todayBonus int64, err error)
}

type FNDailyBonusInitForNewID func() (bonus, dailyBonus int64, err error)

type DailyBonusStorage interface {
	GetAllBonus(id uint64, now time.Time) (bonus, todayBonus int64, err error)
	ConsumeBonus(id uint64, bonusValue, todayBonusValue int64, at time.Time, note string) error
	HasDailyBonus(id uint64, now time.Time) (bool, error)
	EarnDailyBonus(id uint64, now time.Time) error
}
