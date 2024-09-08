package fmstorage

import (
	"path/filepath"
	"sync"
	"time"

	"github.com/godruoyi/go-snowflake"
	"github.com/sgostarter/i/commerr"
	"github.com/sgostarter/i/stg"
	"github.com/sgostarter/libcomponents/trafficpackage"
	"github.com/sgostarter/libeasygo/stg/fs/rawfs"
	"github.com/sgostarter/libeasygo/stg/mwf"
)

func NewFMStorage(root string, storage stg.FileStorage) trafficpackage.Storage {
	return NewFMStorageEx(root, storage, "packages.json", false)
}

func NewFMStorageEx(root string, storage stg.FileStorage, fileName string, prettySerial bool) trafficpackage.Storage {
	if storage == nil {
		storage = rawfs.NewFSStorage("")
	}

	impl := &fmStorageImpl{
		packageStorage: mwf.NewMemWithFile[map[uint64][]*packageD, mwf.Serial, mwf.Lock](
			make(map[uint64][]*packageD), &mwf.JSONSerial{
				MarshalIndent: prettySerial,
			}, &sync.RWMutex{}, filepath.Join(root, fileName), storage),
	}

	return impl
}

type packageD struct {
	trafficpackage.PackageInfo
}

type fmStorageImpl struct {
	packageStorage *mwf.MemWithFile[map[uint64][]*packageD, mwf.Serial, mwf.Lock]
}

func (impl *fmStorageImpl) GetPackageInfo(id, packageID uint64) (info trafficpackage.PackageInfo, err error) {
	impl.packageStorage.Read(func(d map[uint64][]*packageD) {
		pkgs, ok := d[id]
		if !ok {
			err = commerr.ErrNotFound

			return
		}

		for _, pkg := range pkgs {
			if pkg.ID == packageID {
				info = pkg.PackageInfo

				return
			}
		}

		err = commerr.ErrNotFound
	})

	return
}

func (impl *fmStorageImpl) GetPackages(id uint64, includeNoDataPackages bool) (infos []trafficpackage.PackageInfo, err error) {
	impl.packageStorage.Read(func(d map[uint64][]*packageD) {
		pkgs, ok := d[id]
		if !ok {
			return
		}

		for _, pkg := range pkgs {
			if includeNoDataPackages || pkg.LeftAmount > 0 {
				infos = append(infos, pkg.PackageInfo)
			}
		}
	})

	return
}

func (impl *fmStorageImpl) Consume(id uint64, cds []trafficpackage.ConsumeData, at time.Time, note string) error {
	return impl.packageStorage.Change(func(oldD map[uint64][]*packageD) (map[uint64][]*packageD, error) {
		if len(oldD) == 0 {
			oldD = make(map[uint64][]*packageD)
		}

		pkgs, ok := oldD[id]
		if !ok {
			return nil, commerr.ErrNotFound
		}

		fnPackageDByID := func(packageID uint64) *packageD {
			for _, pkg := range pkgs {
				if pkg.ID == packageID {
					return pkg
				}
			}

			return nil
		}

		for _, cd := range cds {
			pkg := fnPackageDByID(cd.PackageID)
			if pkg == nil {
				return nil, commerr.ErrNotFound
			}

			if pkg.LeftAmount != cd.OldAmount {
				return nil, commerr.ErrReject
			}

			if cd.ConsumeAmount > pkg.LeftAmount {
				return nil, commerr.ErrOutOfRange
			}
		}

		for _, cd := range cds {
			pkg := fnPackageDByID(cd.PackageID)

			pkg.LeftAmount -= cd.ConsumeAmount
		}

		return oldD, nil
	})
}

func (impl *fmStorageImpl) AddPackage(id, packageID uint64, amount int64, at time.Time) (newPackageID uint64, err error) {
	if packageID == 0 {
		packageID = snowflake.ID()
	}

	if amount <= 0 {
		return 0, commerr.ErrInvalidArgument
	}

	newPackageID = packageID

	err = impl.packageStorage.Change(func(oldD map[uint64][]*packageD) (map[uint64][]*packageD, error) {
		if len(oldD) == 0 {
			oldD = make(map[uint64][]*packageD)
		}

		pkgs, ok := oldD[id]
		if !ok {
			oldD[id] = make([]*packageD, 0, 1)
		}

		for _, pkg := range pkgs {
			if pkg.ID == packageID {
				return nil, commerr.ErrAlreadyExists
			}
		}

		oldD[id] = append(oldD[id], &packageD{
			PackageInfo: trafficpackage.PackageInfo{
				Package: trafficpackage.Package{
					ID:     packageID,
					Amount: amount,
					At:     at,
				},
				LeftAmount: amount,
			},
		})

		return oldD, nil
	})

	return
}
