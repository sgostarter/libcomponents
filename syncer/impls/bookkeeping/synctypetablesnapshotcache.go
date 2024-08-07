package bookkeeping

import (
	"encoding/json"

	"github.com/sgostarter/libcomponents/syncer"
	"github.com/sgostarter/libeasygo/ptl"
)

const (
	TypeTablePluginID = "sync-table-type"
)

func NewSyncTypeTableSnapshotCache(lastData *syncer.SnapshotData) (syncer.SnapshotPluginCache, bool) {
	impl := &snapshotSyncTypeTableStorage{
		incomeTypes:   make(map[string]TypeRow),
		expensesTypes: make(map[string]TypeRow),
	}

	return impl, impl.init(lastData)
}

type snapshotSyncTypeTableStorage struct {
	incomeTypes   map[string]TypeRow
	expensesTypes map[string]TypeRow
}

func (impl *snapshotSyncTypeTableStorage) init(lastData *syncer.SnapshotData) bool {
	if lastData == nil {
		return false
	}

	for _, record := range lastData.PluginRecords {
		if record.ID != TypeTablePluginID {
			continue
		}

		var data TypeTableSnapshotData

		err := json.Unmarshal(record.Records, &data)
		if err != nil {
			break
		}

		for _, incomeType := range data.IncomeTypes {
			impl.incomeTypes[incomeType.ID] = incomeType
		}

		for _, expensesType := range data.ExpensesTypes {
			impl.expensesTypes[expensesType.ID] = expensesType
		}
	}

	return len(impl.incomeTypes) > 0 || len(impl.expensesTypes) > 0
}

func (impl *snapshotSyncTypeTableStorage) GetID() string {
	return TypeTablePluginID
}

func (impl *snapshotSyncTypeTableStorage) getMapForMetaDataType(mdt MetaDataType) map[string]TypeRow {
	switch mdt {
	case MetaDataIncomeTypeID:
		return impl.incomeTypes
	case MetaDataExpensesTypeID:
		return impl.expensesTypes
	}

	return nil
}

func (impl *snapshotSyncTypeTableStorage) ApplyLog(log syncer.Log) error {
	if log.PluginID != TypeTablePluginID {
		return ptl.NewCodeError(ptl.CodeErrNotExists)
	}

	var data TypeTableLogCore

	err := json.Unmarshal(log.Ds, &data)
	if err != nil {
		return ptl.NewCodeError(ptl.CodeErrLogic)
	}

	m := impl.getMapForMetaDataType(data.MetaDataType)
	if m == nil {
		return ptl.NewCodeError(ptl.CodeErrNotExists)
	}

	switch log.OpType {
	case syncer.OpTypeAdd, syncer.OpTypeChange:
		m[log.RecordID] = TypeRow{
			ID:   log.RecordID,
			Data: log.Ds,
		}
	case syncer.OpTypeDel:
		delete(m, log.RecordID)
	default:
		return ptl.NewCodeError(ptl.CodeErrUnknown)
	}

	return nil
}

func (impl *snapshotSyncTypeTableStorage) GetSnapshotData() (json.RawMessage, error) {
	incomeTypes := make([]TypeRow, 0, len(impl.incomeTypes))

	for _, row := range impl.incomeTypes {
		incomeTypes = append(incomeTypes, row)
	}

	expensesTypesTypes := make([]TypeRow, 0, len(impl.expensesTypes))

	for _, row := range impl.expensesTypes {
		expensesTypesTypes = append(expensesTypesTypes, row)
	}

	return TypeTableSnapshotData{
		IncomeTypes:   incomeTypes,
		ExpensesTypes: expensesTypesTypes,
	}.JSONBytes(), nil
}
