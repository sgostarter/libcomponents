package syncer

import (
	"encoding/json"
	"reflect"
	"time"

	"github.com/sgostarter/libeasygo/stg/kv"
)

type OpType int

const (
	OpTypeAdd OpType = iota
	OpTypeDel
	OpTypeChange
	OpTypeSnapshot
)

type Log struct {
	SeqID    string `json:"seq_id"`
	OpType   OpType `json:"op_type,omitempty"`
	RecordID string `json:"record_id"`
	Ds       []byte `json:"ds,omitempty"`

	VersionID    string `json:"version_id,omitempty"`
	NewVersionID string `json:"new_version_id,omitempty"`

	//
	// plugin
	//
	PluginID string `json:"plugin_id,omitempty"`
}

func EqualLog(log1, log2 Log) bool {
	return reflect.DeepEqual(log1, log2)
}

type Syncer interface {
	AppendAddRecordLog(recordID string, data []byte) error
	AppendDelRecordLog(recordID, versionID string) error
	AppendChangeRecordLog(recordID, versionID string, data []byte) error

	AppendPluginLog(modifier func() (Log, error)) error

	GetAllLogs(startSeqID string) ([]Log, error)
}

type InterruptedLog struct {
	Log         Log
	PoolIndex   int
	LogIDOnPool uint64
}

type PluginSnapshotData struct {
	ID      string          `json:"id"`
	Records json.RawMessage `json:"records,omitempty"`
}

type SnapshotData struct {
	Records       []RecordRow          `json:"records,omitempty"`
	PluginRecords []PluginSnapshotData `json:"plugin_records,omitempty"`
}

type Snapshot interface {
	ApplyAddRecordLog(id string, data []byte, newVersionID string) error
	ApplyChangeRecordLog(id string, versionID string, data []byte, newVersionID string) error
	ApplyDelRecordLog(id string, versionID string) error

	ApplyPluginLog(log Log) error

	GetSnapshotData() (*SnapshotData, error)
}

type Storage interface {
	NewLogPool(idx int) (LogPool, error)
	GetKVStorage() kv.Storage2
	NewSnapshot(lastData *SnapshotData) (Snapshot, error)

	PreLog(log Log, poolIndex int, logIDOnPool uint64) error
	AfterLog() error
	GetInterruptedLog() (log InterruptedLog, exists bool, err error)
}

type LogPool interface {
	GetID() int
	Close()

	AddRecordLog(index uint64, log Log) error
	GetRecordLog(index uint64) (Log, bool, error)
	GetLastRecordLog() (index uint64, log Log, exists bool, err error)
	GetRecordLogs(startIndex, endIndex uint64) ([]Log, error)

	SetSnapshot(d *SnapshotData) error
	GetSnapshot() (d *SnapshotData, err error)
}

type UpdateFlag int

const (
	UpdateFlagWaitSync UpdateFlag = iota
	UpdateFlagSyncToServer
	UpdateFlagSyncDone
)

type RecordRow struct {
	ID         string     `json:"id"`
	Version    string     `json:"version"`
	UpdateFlag UpdateFlag `json:"update_flag"`
	Deleted    bool       `json:"deleted"`
	Data       []byte     `json:"data"`
	At         time.Time  `json:"at"`
}

type SnapshotRecordCache interface {
	GetRecord(id string) (rr RecordRow, exists bool, err error)
	SetRecord(rr RecordRow) error
	DelRecord(id string) error

	GetSnapshotData() ([]RecordRow, error)
}

type SnapshotPluginCache interface {
	GetID() string
	ApplyLog(log Log) error
	GetSnapshotData() (json.RawMessage, error)
}

type SnapshotPluginCacheManager interface {
	GetCache(id string) (SnapshotPluginCache, error)
	GetCaches4Save() ([]SnapshotPluginCache, error)
}
