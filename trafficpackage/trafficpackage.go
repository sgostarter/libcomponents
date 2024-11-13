package trafficpackage

import (
	"time"

	"github.com/sgostarter/i/commerr"
	"github.com/sgostarter/i/l"
)

func NewTrafficPackage(stableID string, storage Storage, logger l.Wrapper) TrafficPackage {
	if logger == nil {
		logger = l.NewNopLoggerWrapper()
	}

	logger = logger.WithFields(l.StringField(l.ClsKey, "trafficPackageImpl"))

	if storage == nil {
		logger.Fatal("no storage")
	}

	return &trafficPackageImpl{
		logger:   logger,
		stableID: stableID,
		storage:  storage,
	}
}

type trafficPackageImpl struct {
	logger   l.Wrapper
	stableID string
	storage  Storage
}

func (impl *trafficPackageImpl) GetPackageInfo(id, packageID uint64) (PackageInfo, error) {
	return impl.storage.GetPackageInfo(id, packageID)
}

func (impl *trafficPackageImpl) AddPackage(id uint64, amount int64, at time.Time) (newPackageID uint64, err error) {
	return impl.AddPackageEx(id, 0, amount, at)
}

func (impl *trafficPackageImpl) AddPackageEx(id, packageID uint64, amount int64, at time.Time) (newPackageID uint64, err error) {
	return impl.storage.AddPackage(id, packageID, amount, at)
}

func (impl *trafficPackageImpl) GetAmount(id uint64) (amount int64, err error) {
	packageInfos, err := impl.storage.GetPackages(id, false)
	if err != nil {
		return
	}

	for _, info := range packageInfos {
		amount += info.LeftAmount
	}

	return
}

func (impl *trafficPackageImpl) ConsumeAmount(id uint64, _ time.Time, n int64, at time.Time, note string) (err error) {
	if n <= 0 {
		return
	}

	packageInfos, err := impl.storage.GetPackages(id, false)
	if err != nil {
		return
	}

	cds := make([]ConsumeData, 0, len(packageInfos))

	for _, info := range packageInfos {
		if n <= info.LeftAmount {
			cds = append(cds, ConsumeData{
				PackageID:     info.ID,
				ConsumeAmount: n,
				OldAmount:     info.LeftAmount,
			})

			n = 0

			break
		}

		cds = append(cds, ConsumeData{
			PackageID:     info.ID,
			ConsumeAmount: info.LeftAmount,
			OldAmount:     info.LeftAmount,
		})

		n -= info.LeftAmount
	}

	if n > 0 {
		err = commerr.ErrOutOfRange

		return
	}

	err = impl.storage.Consume(id, cds, at, note)

	return
}

func (impl *trafficPackageImpl) GetStableID() string {
	return impl.stableID
}

func (impl *trafficPackageImpl) TryConsumeAmount(id uint64, _ time.Time, n int64, at time.Time, note string) (rn int64, err error) {
	if n <= 0 {
		return
	}

	rn = n

	packageInfos, err := impl.storage.GetPackages(id, false)
	if err != nil {
		return
	}

	for _, info := range packageInfos {
		if n <= info.LeftAmount {
			err = impl.storage.Consume(id, []ConsumeData{
				{
					PackageID:     info.ID,
					ConsumeAmount: n,
					OldAmount:     info.LeftAmount,
				},
			}, at, note)

			if err == nil {
				n = 0
			} else {
				if n != rn {
					err = nil
				}
			}

			break
		}

		err = impl.storage.Consume(id, []ConsumeData{
			{
				PackageID:     info.ID,
				ConsumeAmount: info.LeftAmount,
				OldAmount:     info.LeftAmount,
			},
		}, at, note)

		if err != nil {
			if n != rn {
				err = nil
			}

			break
		}

		n -= info.LeftAmount
	}

	rn -= n

	return
}

func (impl *trafficPackageImpl) GetPackages(id uint64, includeNoDataPackages bool) ([]PackageInfo, error) {
	return impl.storage.GetPackages(id, includeNoDataPackages)
}
