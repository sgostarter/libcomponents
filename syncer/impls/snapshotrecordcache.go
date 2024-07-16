package impls

import (
	"slices"

	"github.com/sgostarter/libcomponents/syncer"
)

func NewSnapshotRecordCache(lastData *syncer.SnapshotData) syncer.SnapshotRecordCache {
	impl := &snapshotStorageImpl{
		records: make(map[string]syncer.RecordRow),
	}

	impl.init(lastData)

	return impl
}

type snapshotStorageImpl struct {
	records map[string]syncer.RecordRow
}

func (impl *snapshotStorageImpl) init(lastData *syncer.SnapshotData) {
	if lastData == nil {
		return
	}

	for _, record := range lastData.Records {
		impl.records[record.ID] = record
	}
}

func (impl *snapshotStorageImpl) GetRecord(id string) (rr syncer.RecordRow, exists bool, err error) {
	rr, exists = impl.records[id]

	return
}

func (impl *snapshotStorageImpl) SetRecord(rr syncer.RecordRow) error {
	impl.records[rr.ID] = rr

	return nil
}

func (impl *snapshotStorageImpl) DelRecord(id string) error {
	delete(impl.records, id)

	return nil
}

func (impl *snapshotStorageImpl) GetSnapshotData() ([]syncer.RecordRow, error) {
	records := make([]syncer.RecordRow, 0, len(impl.records))

	for _, row := range impl.records {
		records = append(records, row)
	}

	slices.SortFunc(records, func(a, b syncer.RecordRow) int {
		if a.At == b.At {
			return 0
		}

		if a.At < b.At {
			return -1
		}

		return 1
	})

	return records, nil
}
