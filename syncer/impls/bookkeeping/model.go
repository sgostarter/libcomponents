package bookkeeping

import (
	"encoding/json"

	"github.com/sgostarter/libeasygo/ptl"
)

const (
	CodeErrLabelExists = ptl.CodeErrCustomStart + iota + 1
)

type TypeTable interface {
	Reset()

	TestAdd(id, label, parentID string) (ok bool, err error)
	Add(id, label, parentID string, data []byte) error
	TestDel(id, toRecordID string) (ok bool, err error)
	Del(id, toRecordID string) error
	TestChange(id, label, parentID string) (ok bool, err error)
	Change(id, label, parentID string, data []byte) error
}

type TypeRow struct {
	ID   string `json:"id"`
	Data []byte `json:"data"`
}

type MetaDataType int

const (
	MetaDataIncomeTypeID MetaDataType = iota
	MetaDataExpensesTypeID
)

type TypeTableLogCore struct {
	MetaDataType MetaDataType `json:"meta_data_type,omitempty"`
}

func (ttl TypeTableLogCore) JSONBytes() json.RawMessage {
	d, err := json.Marshal(ttl)
	if err != nil {
		return nil
	}

	return d
}

type TypeTableSnapshotData struct {
	IncomeTypes   []TypeRow `json:"income_types,omitempty"`
	ExpensesTypes []TypeRow `json:"expenses_types,omitempty"`
}

func (ttd TypeTableSnapshotData) JSONBytes() json.RawMessage {
	d, err := json.Marshal(ttd)
	if err != nil {
		return nil
	}

	return d
}
