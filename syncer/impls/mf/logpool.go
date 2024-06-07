package mf

import (
	"encoding/json"
	"os"
	"sync"

	"github.com/sgostarter/i/stg"
	"github.com/sgostarter/libcomponents/syncer"
	"github.com/sgostarter/libeasygo/ptl"
	"github.com/sgostarter/libeasygo/stg/fs/rawfs"
	"github.com/sgostarter/libeasygo/stg/mwf"
)

func NewMFLogPoolEx(id int, file string, storage stg.FileStorage) syncer.LogPool {
	if storage == nil {
		storage = rawfs.NewFSStorage("")
	}

	return &logPoolImpl{
		id:      id,
		file:    file,
		storage: storage,
		d: mwf.NewMemWithFile[[]syncer.Log, mwf.Serial, mwf.Lock](make([]syncer.Log, 0),
			&mwf.JSONSerial{}, &sync.RWMutex{}, file, storage),
	}
}

type logPoolImpl struct {
	id      int
	file    string
	storage stg.FileStorage
	d       *mwf.MemWithFile[[]syncer.Log, mwf.Serial, mwf.Lock]
}

func (impl *logPoolImpl) GetID() int {
	return impl.id
}

func (impl *logPoolImpl) Close() {

}

func (impl *logPoolImpl) AddRecordLog(index uint64, log syncer.Log) error {
	return impl.d.Change(func(v []syncer.Log) (newV []syncer.Log, err error) {
		newV = v

		if newV == nil {
			newV = make([]syncer.Log, 0, 10)
		}

		if len(newV) != int(index) {
			err = ptl.NewCodeError(ptl.CodeErrLogic)

			return
		}

		log.SeqID = index

		newV = append(newV, log)

		return
	})
}

func (impl *logPoolImpl) GetRecordLog(index uint64) (log syncer.Log, exists bool, err error) {
	logs, err := impl.GetRecordLogs(index, index+1)
	if err != nil {
		return
	}

	if len(logs) == 0 {
		return
	}

	exists = true
	log = logs[0]

	return
}

func (impl *logPoolImpl) GetLastRecordLog() (index uint64, log syncer.Log, exists bool, err error) {
	impl.d.Read(func(v []syncer.Log) {
		l := uint64(len(v))
		if l == 0 {
			return
		}

		index = uint64(len(v) - 1)
		log = v[index]
		exists = true
	})

	return
}

func (impl *logPoolImpl) GetRecordLogs(startIndex, endIndex uint64) (logs []syncer.Log, _ error) {
	impl.d.Read(func(v []syncer.Log) {
		l := uint64(len(v))
		if l == 0 {
			return
		}

		if endIndex == 0 || endIndex > l {
			endIndex = l
		}

		if endIndex < startIndex {
			return
		}

		logs = append(logs, v[int(startIndex):int(endIndex)]...)
	})

	return
}

func (impl *logPoolImpl) SetSnapshot(data *syncer.SnapshotData) (err error) {
	d, err := json.MarshalIndent(&data, "", "  ")
	if err != nil {
		return
	}

	err = impl.storage.WriteFile(impl.file+".snapshot", d)

	return
}

func (impl *logPoolImpl) GetSnapshot() (data *syncer.SnapshotData, err error) {
	d, err := impl.storage.ReadFile(impl.file + ".snapshot")
	if os.IsNotExist(err) {
		err = nil

		return
	}

	if err != nil {
		return
	}

	data = &syncer.SnapshotData{}

	err = json.Unmarshal(d, data)

	return
}
