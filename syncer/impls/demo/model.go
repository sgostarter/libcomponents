package demo

import (
	"encoding/json"
	"time"

	"github.com/sgostarter/libeasygo/ptl"
)

const (
	CodeErrLabelExists = ptl.CodeErrCustomStart + iota + 1
)

type TypeTable interface {
	Add(id, label, parentID string, data []byte) error
	Del(id string) error
	Change(id, label, parentID string, data []byte) error
}

type TypeRow struct {
	ID       string    `json:"id"`
	Label    string    `json:"label"`
	Data     []byte    `json:"data"`
	ParentID string    `json:"parent_id"`
	ToID     string    `json:"to_id"`
	At       time.Time `json:"at"`
}

type MetaDataType int

const (
	MetaDataIncomeTypeID MetaDataType = iota
	MetaDataExpensesTypeID
)

type TypeTableLog struct {
	MetaDataType MetaDataType `json:"meta_data_type,omitempty"`
	Label        string       `json:"label,omitempty"`
	ParentID     string       `json:"parent_id,omitempty"`
	ToRecordID   string       `json:"to_record_id,omitempty"`
	At           time.Time    `json:"at,omitempty"`
}

func (ttl TypeTableLog) JSONBytes() json.RawMessage {
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
