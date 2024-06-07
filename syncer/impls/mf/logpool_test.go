// nolint
package mf

import (
	"os"
	"reflect"
	"testing"

	"github.com/sgostarter/libcomponents/syncer"
	"github.com/sgostarter/libeasygo/pathutils"
	"github.com/sgostarter/libeasygo/stg/fs/rawfs"
	"github.com/stretchr/testify/assert"
)

const (
	utRoot = "ut-data"
)

func TestMain(m *testing.M) {
	_ = os.RemoveAll(utRoot)
	_ = pathutils.MustDirExists(utRoot)

	code := m.Run()

	_ = os.RemoveAll(utRoot)

	os.Exit(code)
}

func TestLogPool(t *testing.T) {
	_ = os.RemoveAll(utRoot)
	_ = pathutils.MustDirExists(utRoot)

	lp := NewMFLogPoolEx(10, "log-pool-test", rawfs.NewFSStorage(utRoot))

	_, exists, err := lp.GetRecordLog(0)
	assert.Nil(t, err)
	assert.False(t, exists)

	_, exists, err = lp.GetRecordLog(1)
	assert.Nil(t, err)
	assert.False(t, exists)

	log1 := syncer.Log{
		OpType:   syncer.OpTypeAdd,
		RecordID: "1",
		Ds: [][]byte{
			[]byte("111"),
		},
		NewVersionID: "1",
	}

	err = lp.AddRecordLog(1, log1)
	assert.NotNil(t, err)

	err = lp.AddRecordLog(0, log1)
	assert.Nil(t, err)

	rLog1, exists, err := lp.GetRecordLog(0)
	assert.Nil(t, err)
	assert.True(t, exists)
	assert.True(t, reflect.DeepEqual(log1, rLog1))

	_, exists, err = lp.GetRecordLog(1)
	assert.Nil(t, err)
	assert.False(t, exists)

	log2 := syncer.Log{
		OpType:   syncer.OpTypeAdd,
		RecordID: "2",
		Ds: [][]byte{
			[]byte("222"),
		},
		NewVersionID: "2",
	}

	err = lp.AddRecordLog(1, log2)

	log2.SeqID = 1

	assert.Nil(t, err)

	rLog1, exists, err = lp.GetRecordLog(0)
	assert.Nil(t, err)
	assert.True(t, exists)
	assert.True(t, reflect.DeepEqual(log1, rLog1))

	rLog2, exists, err := lp.GetRecordLog(1)
	assert.Nil(t, err)
	assert.True(t, exists)
	assert.True(t, reflect.DeepEqual(log2, rLog2))

	logs, err := lp.GetRecordLogs(0, 1)
	assert.Nil(t, err)
	assert.True(t, len(logs) == 1)

	logs, err = lp.GetRecordLogs(0, 2)
	assert.Nil(t, err)
	assert.True(t, len(logs) == 2)

	logs, err = lp.GetRecordLogs(0, 3)
	assert.Nil(t, err)
	assert.True(t, len(logs) == 2)

	logs, err = lp.GetRecordLogs(1, 3)
	assert.Nil(t, err)
	assert.True(t, len(logs) == 1)
}
