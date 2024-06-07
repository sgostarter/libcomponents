package impls

import (
	"encoding/json"

	"github.com/sgostarter/i/l"
	"github.com/sgostarter/libcomponents/syncer"
	"github.com/sgostarter/libeasygo/ptl"
)

func NewSnapshot(recordCache syncer.SnapshotRecordCache, pluginStoreManager syncer.SnapshotPluginCacheManager, logger l.Wrapper) syncer.Snapshot {
	if logger == nil {
		logger = l.NewNopLoggerWrapper()
	}

	logger = logger.WithFields(l.StringField(l.ClsKey, "snapshotImpl"))

	if recordCache == nil {
		logger.Fatal("no recordCache")
	}

	return &snapshotImpl{
		logger:             logger,
		recordCache:        recordCache,
		pluginStoreManager: pluginStoreManager,
	}
}

type snapshotImpl struct {
	logger l.Wrapper

	recordCache        syncer.SnapshotRecordCache
	pluginStoreManager syncer.SnapshotPluginCacheManager
}

func (impl *snapshotImpl) ApplyAddRecordLog(id string, data []byte, newVersionID string) (err error) {
	rr, exists, err := impl.recordCache.GetRecord(id)
	if err != nil {
		impl.logger.WithFields(l.ErrorField(err), l.StringField("id", id)).
			Error("apply add record: get record failed")

		return
	}

	if exists {
		if rr.UpdateFlag == syncer.UpdateFlagSyncDone {
			err = ptl.NewCodeError(ptl.CodeErrExists)

			impl.logger.WithFields(l.StringField("id", id)).
				Error("apply add record: update flag is UpdateFlagSyncDone")

			return
		}
	}

	err = impl.recordCache.SetRecord(syncer.RecordRow{
		ID:         id,
		Version:    newVersionID,
		UpdateFlag: syncer.UpdateFlagSyncDone,
		Data:       append([]byte{}, data...),
	})

	if err != nil {
		impl.logger.WithFields(l.ErrorField(err), l.StringField("id", id)).
			Error("apply add record: set record failed")

		return
	}

	return
}

func (impl *snapshotImpl) ApplyChangeRecordLog(id string, versionID string, data []byte, newVersionID string) (err error) {
	rr, exists, err := impl.recordCache.GetRecord(id)
	if err != nil {
		impl.logger.WithFields(l.ErrorField(err), l.StringField("id", id)).
			Error("apply change record: get record failed")

		return
	}

	if !exists {
		err = ptl.NewCodeError(ptl.CodeErrNotExists)

		impl.logger.WithFields(l.StringField("id", id)).Error("apply change record: no record")

		return
	}

	if rr.UpdateFlag == syncer.UpdateFlagSyncDone && rr.Version != versionID {
		err = ptl.NewCodeError(ptl.CodeErrConflict)

		impl.logger.WithFields(l.StringField("id", id), l.StringField("version", versionID),
			l.StringField("dbVersion", rr.Version)).Error("apply change record: version mismatch")

		return
	}

	err = impl.recordCache.SetRecord(syncer.RecordRow{
		ID:         id,
		Version:    newVersionID,
		UpdateFlag: syncer.UpdateFlagSyncDone,
		Data:       append([]byte{}, data...),
	})

	if err != nil {
		impl.logger.WithFields(l.ErrorField(err), l.StringField("id", id)).
			Error("apply change record: set record failed")

		return
	}

	return
}

func (impl *snapshotImpl) ApplyDelRecordLog(id string, versionID string) (err error) {
	rr, exists, err := impl.recordCache.GetRecord(id)
	if err != nil {
		impl.logger.WithFields(l.ErrorField(err), l.StringField("id", id)).
			Error("apply del record: get record failed")

		return
	}

	if !exists {
		err = ptl.NewCodeError(ptl.CodeErrNotExists)

		impl.logger.WithFields(l.StringField("id", id)).Error("apply del record: no record")

		return
	}

	if rr.UpdateFlag == syncer.UpdateFlagSyncDone && rr.Version != versionID {
		err = ptl.NewCodeError(ptl.CodeErrConflict)

		impl.logger.WithFields(l.StringField("id", id), l.StringField("version", versionID),
			l.StringField("dbVersion", rr.Version)).Error("apply del record: version mismatch")

		return
	}

	err = impl.recordCache.DelRecord(id)
	if err != nil {
		impl.logger.WithFields(l.ErrorField(err), l.StringField("id", id)).
			Error("apply del record: del record failed")

		return
	}

	return
}

func (impl *snapshotImpl) ApplyPluginLog(log syncer.Log) error {
	if impl.pluginStoreManager == nil {
		impl.logger.WithFields(l.StringField("pluginID", log.PluginID)).
			Error("apply plugin log: no plugin manager")

		return ptl.NewCodeError(ptl.CodeErrNotExists)
	}

	pluginStore, err := impl.pluginStoreManager.GetCache(log.PluginID)
	if err != nil {
		impl.logger.WithFields(l.ErrorField(err), l.StringField("pluginID", log.PluginID)).
			Error("apply plugin log: get plugin failed")

		return ptl.NewCodeError(ptl.CodeErrNotExists)
	}

	if pluginStore == nil {
		impl.logger.WithFields(l.StringField("pluginID", log.PluginID)).
			Error("apply plugin log: no plugin")

		return ptl.NewCodeError(ptl.CodeErrNotExists)
	}

	err = pluginStore.ApplyLog(log)
	if err != nil {
		impl.logger.WithFields(l.ErrorField(err), l.StringField("id", log.RecordID)).
			Error("apply other log: add other failed")

		return err
	}

	return nil
}

func (impl *snapshotImpl) getPluginSnapshotDatas() (records []syncer.PluginSnapshotData, err error) {
	if impl.pluginStoreManager == nil {
		return
	}

	pluginStores, err := impl.pluginStoreManager.GetCaches4Save()
	if err != nil {
		return
	}

	records = make([]syncer.PluginSnapshotData, 0, len(pluginStores))

	for _, store := range pluginStores {
		var pluginRecords json.RawMessage

		pluginRecords, err = store.GetSnapshotData()
		if err != nil {
			return
		}

		records = append(records, syncer.PluginSnapshotData{
			ID:      store.GetID(),
			Records: pluginRecords,
		})
	}

	return
}

func (impl *snapshotImpl) GetSnapshotData() (data *syncer.SnapshotData, err error) {
	records, err := impl.recordCache.GetSnapshotData()
	if err != nil {
		return
	}

	pluginRecords, err := impl.getPluginSnapshotDatas()
	if err != nil {
		return
	}

	data = &syncer.SnapshotData{
		Records:       records,
		PluginRecords: pluginRecords,
	}

	return
}
