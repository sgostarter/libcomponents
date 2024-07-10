package syncer

import (
	"github.com/sgostarter/i/l"
	"github.com/sgostarter/libeasygo/ptl"
)

func (impl *syncerImpl) trySnapshot(logPoolIndex int) {
	if impl.snapshotLogCount <= 0 {
		return
	}

	select {
	case impl.chSnapshotBuild <- logPoolIndex:
	default:
	}
}

func (impl *syncerImpl) GetLastSnapshotData(logPoolIndex int) (d *SnapshotData, err error) {
	if logPoolIndex <= 0 {
		return
	}

	logPool, err := impl.store.NewLogPool(logPoolIndex - 1)
	if err != nil {
		return
	}

	d, err = logPool.GetSnapshot()

	return
}

// nolint: funlen
func (impl *syncerImpl) buildSnapshotOnRoutine(logPoolIndex int, logger l.Wrapper) {
	logPool, err := impl.store.NewLogPool(logPoolIndex)
	if err != nil {
		logger.WithFields(l.ErrorField(err), l.IntField("index", logPoolIndex)).Error("get log pool failed")

		return
	}

	snapshotData, err := logPool.GetSnapshot()
	if err != nil {
		logger.WithFields(l.ErrorField(err), l.IntField("index", logPoolIndex)).Error("get snap shot failed")

		return
	}

	if snapshotData != nil {
		return
	}

	logs, err := logPool.GetRecordLogs(0, 0)
	if err != nil {
		logger.WithFields(l.ErrorField(err), l.IntField("index", logPoolIndex)).Error("get records failed")

		return
	}

	lastSnapshotData, err := impl.GetLastSnapshotData(logPoolIndex)
	if err != nil {
		logger.WithFields(l.ErrorField(err), l.IntField("index", logPoolIndex)).Error("get last snapshot data failed")

		return
	}

	snapshot, err := impl.store.NewSnapshot(lastSnapshotData)
	if err != nil {
		logger.WithFields(l.ErrorField(err), l.IntField("index", logPoolIndex)).Error("get snap shot failed")

		return
	}

	for _, log := range logs {
		if log.PluginID == "" {
			switch log.OpType {
			case OpTypeAdd:
				err = snapshot.ApplyAddRecordLog(log.RecordID, log.Ds, log.NewVersionID)
			case OpTypeDel:
				err = snapshot.ApplyDelRecordLog(log.RecordID, log.VersionID)
			case OpTypeChange:
				err = snapshot.ApplyChangeRecordLog(log.RecordID, log.VersionID, log.Ds, log.NewVersionID)
			default:
				err = ptl.NewCodeError(ptl.CodeErrLogic)
			}
		} else {
			err = snapshot.ApplyPluginLog(log)
		}

		if err != nil {
			logger.WithFields(l.ErrorField(err), l.UInt64Field("seqID", log.SeqID)).Error("process log failed")
		}
	}

	ds, err := snapshot.GetSnapshotData()
	if err != nil {
		logger.WithFields(l.ErrorField(err), l.IntField("index", logPoolIndex)).Error("get snap shot data failed")

		return
	}

	err = logPool.SetSnapshot(ds)
	if err != nil {
		logger.WithFields(l.ErrorField(err), l.IntField("index", logPoolIndex)).Error("set snap shot failed")

		return
	}
}

func (impl *syncerImpl) snapShotBuildRoutine() {
	logger := impl.logger.WithFields(l.StringField(l.RoutineKey, "syncerImpl"))

	logger.Debug("enter")

	defer logger.Debug("leave")

	defer impl.wg.Done()

	loop := true

	for loop {
		select {
		case <-impl.ctx.Done():
			loop = false

			break
		case poolIndex := <-impl.chSnapshotBuild:
			impl.buildSnapshotOnRoutine(poolIndex, logger)
		}
	}
}
