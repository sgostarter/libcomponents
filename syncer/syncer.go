package syncer

import (
	"context"
	"encoding/json"
	"strconv"
	"sync"

	"github.com/godruoyi/go-snowflake"
	"github.com/sgostarter/i/l"
	"github.com/spf13/cast"
)

const (
	kvCurLogPoolKey             = "curLogPool"
	kvNextLogIDOnCurrentPoolKey = "nextLogIDOnCurrentPool"
)

func NewSyncer(ctx context.Context, store Storage, snapshotLogCount uint64, logger l.Wrapper) Syncer {
	if logger == nil {
		logger = l.NewNopLoggerWrapper()
	}

	logger = logger.WithFields(l.StringField(l.ClsKey, "syncerImpl"))

	if store == nil {
		logger.Fatalf("no store")
	}

	ctx, cancel := context.WithCancel(ctx)

	impl := &syncerImpl{
		ctx:              ctx,
		ctxCancel:        cancel,
		logger:           logger,
		store:            store,
		snapshotLogCount: snapshotLogCount,
		chSnapshotBuild:  make(chan int, 10),
	}

	impl.init()

	return impl
}

type syncerImpl struct {
	ctx       context.Context
	ctxCancel context.CancelFunc
	wg        sync.WaitGroup

	logger l.Wrapper
	store  Storage

	snapshotLogCount uint64

	poolLock               sync.Mutex
	currentLogPoolIndex    int
	nextLogIDOnCurrentPool uint64

	logPool LogPool

	chSnapshotBuild chan int
}

func (impl *syncerImpl) init() {
	vs, err := impl.store.GetKVStorage().Gets([]string{kvCurLogPoolKey, kvNextLogIDOnCurrentPoolKey})
	if err != nil {
		impl.logger.WithFields(l.ErrorField(err)).Fatal("startup failed")
	}

	impl.currentLogPoolIndex = cast.ToInt(vs[0])
	impl.nextLogIDOnCurrentPool = cast.ToUint64(vs[1])

	impl.logPool, err = impl.store.NewLogPool(impl.currentLogPoolIndex)
	if err != nil {
		impl.logger.WithFields(l.ErrorField(err)).Fatal("get log pool")
	}

	impl.wg.Add(1)
	go impl.snapShotBuildRoutine()
}

func (impl *syncerImpl) getPoolIndex() (poolIndex int, logIDonPool uint64) {
	impl.poolLock.Lock()
	defer impl.poolLock.Unlock()

	return impl.getPoolIndexOnLock()
}

func (impl *syncerImpl) getPoolIndexOnLock() (poolIndex int, logIDonPool uint64) {
	if impl.snapshotLogCount <= 0 {
		return 0, impl.nextLogIDOnCurrentPool
	}

	if impl.nextLogIDOnCurrentPool < impl.snapshotLogCount {
		return impl.currentLogPoolIndex, impl.nextLogIDOnCurrentPool
	}

	return impl.currentLogPoolIndex + 1, 0
}

func (impl *syncerImpl) savePoolIndexOnLock(poolIndex int, logIDonPool uint64) error {
	if err := impl.store.GetKVStorage().Sets([]string{kvCurLogPoolKey, kvNextLogIDOnCurrentPoolKey},
		poolIndex, logIDonPool); err != nil {
		return err
	}

	impl.currentLogPoolIndex = poolIndex
	impl.nextLogIDOnCurrentPool = logIDonPool

	return nil
}

func (impl *syncerImpl) mustLogPoolByIndexOnLock(poolIndex int) (err error) {
	if impl.logPool != nil && impl.logPool.GetID() != poolIndex {
		impl.logPool.Close()

		impl.trySnapshot(impl.logPool.GetID())

		impl.logPool = nil
	}

	if impl.logPool == nil {
		impl.logPool, err = impl.store.NewLogPool(poolIndex)
		if err != nil {
			return
		}
	}

	return
}

func (impl *syncerImpl) processUnexpectedLogsOnLock() (err error) { // TODO check this logic on ut
	log, exists, err := impl.store.GetInterruptedLog()
	if err != nil {
		return
	}

	if !exists {
		return
	}

	poolIndex, _ := impl.getPoolIndexOnLock()
	if log.PoolIndex != poolIndex {
		_ = impl.store.AfterLog()

		return
	}

	err = impl.mustLogPoolByIndexOnLock(log.PoolIndex)
	if err != nil {
		return
	}

	//
	lastLogIndexOnPool, dbLog, exists, err := impl.logPool.GetLastRecordLog()
	if err != nil {
		return
	}

	if !exists {
		_ = impl.store.AfterLog()

		return
	}

	if lastLogIndexOnPool != log.LogIDOnPool {
		_ = impl.store.AfterLog()

		return
	}

	if EqualLog(dbLog, log.Log) {
		err = impl.savePoolIndexOnLock(log.PoolIndex, log.LogIDOnPool+1)
		if err != nil {
			return
		}
	}

	_ = impl.store.AfterLog()

	return
}

func (impl *syncerImpl) tryNewLog(cb func() (Log, error)) (err error) {
	impl.poolLock.Lock()
	defer impl.poolLock.Unlock()

	err = impl.processUnexpectedLogsOnLock()
	if err != nil {
		return
	}

	log, err := cb()
	if err != nil {
		return
	}

	poolIndex, logIndexOnPool := impl.getPoolIndexOnLock()

	err = impl.mustLogPoolByIndexOnLock(poolIndex)
	if err != nil {
		return
	}

	err = impl.store.PreLog(log, poolIndex, logIndexOnPool)
	if err != nil {
		return
	}

	err = impl.logPool.AddRecordLog(logIndexOnPool, log)
	if err != nil {
		_ = impl.store.AfterLog()

		return
	}

	err = impl.savePoolIndexOnLock(poolIndex, logIndexOnPool+1)
	if err != nil {
		return
	}

	_ = impl.store.AfterLog()

	return
}

func (impl *syncerImpl) newVersionID() string {
	return strconv.FormatUint(snowflake.ID(), 36)
}

func (impl *syncerImpl) AppendAddRecordLog(recordID string, data []byte) error {
	return impl.tryNewLog(func() (Log, error) {
		return Log{
			OpType:       OpTypeAdd,
			RecordID:     recordID,
			Ds:           data,
			NewVersionID: impl.newVersionID(),
		}, nil
	})
}

func (impl *syncerImpl) AppendDelRecordLog(recordID, versionID string) error {
	return impl.tryNewLog(func() (Log, error) {
		return Log{
			OpType:    OpTypeDel,
			RecordID:  recordID,
			VersionID: versionID,
		}, nil
	})
}

func (impl *syncerImpl) AppendChangeRecordLog(recordID, versionID string, data []byte) error {
	return impl.tryNewLog(func() (Log, error) {
		return Log{
			OpType:       OpTypeChange,
			RecordID:     recordID,
			VersionID:    versionID,
			Ds:           data,
			NewVersionID: impl.newVersionID(),
		}, nil
	})
}

func (impl *syncerImpl) AppendPluginLog(modifier func() (Log, error)) error {
	return impl.tryNewLog(func() (Log, error) {
		return modifier()
	})
}

func (impl *syncerImpl) GetAllLogs(startSeqID string) (logs []Log, err error) {
	var startPoolIndex int

	var startLogIndexOnPool uint64

	var startSeqIDN uint64

	if startSeqID == "" {
		startSeqIDN = 0
	} else {
		startSeqIDN, err = SeqIDS2N(startSeqID)
		if err != nil {
			impl.logger.WithFields(l.ErrorField(err), l.StringField("startSeqID", startSeqID)).Error("invalid start seq id")

			return
		}

		startSeqIDN++
	}

	if impl.snapshotLogCount > 0 {
		startPoolIndex = int(startSeqIDN / impl.snapshotLogCount)
		startLogIndexOnPool = startSeqIDN - uint64(startPoolIndex)*impl.snapshotLogCount
	} else {
		startLogIndexOnPool = startSeqIDN
	}

	logs = make([]Log, 0, 100)

	lastPoolIndex, logIDonPool := impl.getPoolIndex()
	if lastPoolIndex == 0 && logIDonPool == 0 {
		return
	}

	if logIDonPool <= 0 {
		lastPoolIndex--
	}

	// use snapshot: (start id is '') && (lastPoolIndex > 0 && snapshot(lastPoolIndex) exists)

	if startSeqIDN == 0 && lastPoolIndex > 0 {
		var d *SnapshotData

		d, err = impl.GetLastSnapshotData(lastPoolIndex)
		if err != nil {
			impl.logger.WithFields(l.ErrorField(err), l.IntField("lastPoolIndex", lastPoolIndex)).Error("get last snapshot data failed")

			return
		}

		if d != nil {
			var data []byte

			data, err = json.Marshal(d)
			if err != nil {
				impl.logger.WithFields(l.ErrorField(err)).Error("marshal snapshot data failed")

				return
			}

			logs = append(logs, Log{
				SeqID:  uint64(lastPoolIndex)*impl.snapshotLogCount - 1,
				OpType: OpTypeSnapshot,
				Ds:     data,
			})

			startPoolIndex = lastPoolIndex
			startLogIndexOnPool = 0
		}
	}

	for ; startPoolIndex < lastPoolIndex; startPoolIndex++ {
		logPool, e := impl.store.NewLogPool(startPoolIndex)
		if e != nil {
			return
		}

		poolLogs, e := logPool.GetRecordLogs(startLogIndexOnPool, 0)
		if e != nil {
			return
		}

		for idx, log := range poolLogs {
			log.SeqID = uint64(startPoolIndex)*impl.snapshotLogCount + uint64(idx) + startLogIndexOnPool

			logs = append(logs, log)
		}

		startLogIndexOnPool = 0
	}

	logPool, err := impl.store.NewLogPool(startPoolIndex)
	if err != nil {
		return
	}

	poolLogs, err := logPool.GetRecordLogs(startLogIndexOnPool, 0)
	if err != nil {
		return
	}

	for idx, log := range poolLogs {
		log.SeqID = uint64(startPoolIndex)*impl.snapshotLogCount + uint64(idx) + startLogIndexOnPool

		logs = append(logs, log)
	}

	return
}
