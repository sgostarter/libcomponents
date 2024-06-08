package onlyrecord

import (
	"fmt"
	"path"

	"github.com/sgostarter/i/l"
	"github.com/sgostarter/libcomponents/syncer"
	"github.com/sgostarter/libcomponents/syncer/impls"
	"github.com/sgostarter/libcomponents/syncer/impls/mf"
	"github.com/sgostarter/libeasygo/pathutils"
	"github.com/sgostarter/libeasygo/stg/fs/rawfs"
	"github.com/sgostarter/libeasygo/stg/kv"
	"github.com/sgostarter/libeasygo/stg/mwf"
)

func NewMFStorage(dataRoot string, logger l.Wrapper) syncer.Storage {
	if logger == nil {
		logger = l.NewNopLoggerWrapper()
	}

	logger = logger.WithFields(l.StringField(l.ClsKey, "storageImpl"))

	impl := &storageImpl{
		logger:   logger,
		dataRoot: dataRoot,
	}

	impl.init()

	return impl
}

type storageImpl struct {
	logger   l.Wrapper
	dataRoot string

	kv kv.StorageTiny2
}

func (impl *storageImpl) init() {
	_ = pathutils.MustDirExists(impl.dataRoot)

	impl.kv = mwf.NewKVEx("kv.dat", rawfs.NewFSStorage(impl.dataRoot))
}

func (impl *storageImpl) NewLogPool(idx int) (syncer.LogPool, error) {
	_ = pathutils.MustDirExists(impl.dataRoot)

	return mf.NewMFLogPoolEx(idx, path.Join(impl.dataRoot, fmt.Sprintf("log-pool_%d.dat", idx)), nil), nil
}

func (impl *storageImpl) GetKVStorage() kv.Storage2 {
	return impl.kv
}

func (impl *storageImpl) NewSnapshot(lastData *syncer.SnapshotData) (syncer.Snapshot, error) {
	return impls.NewSnapshot(impls.NewSnapshotRecordCache(lastData), nil, impl.logger), nil
}

func (impl *storageImpl) PreLog(log syncer.Log, poolIndex int, logIDOnPool uint64) error {
	return impl.kv.Set(":recover-log", syncer.InterruptedLog{
		Log:         log,
		PoolIndex:   poolIndex,
		LogIDOnPool: logIDOnPool,
	})
}

func (impl *storageImpl) AfterLog() error {
	return impl.kv.Del(":recover-log")
}

func (impl *storageImpl) GetInterruptedLog() (log syncer.InterruptedLog, exists bool, err error) {
	exists, err = impl.kv.Get(":recover-log", &log)

	return
}
