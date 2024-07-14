// nolint
package bookeepingtest

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/sgostarter/i/l"
	"github.com/sgostarter/libcomponents/syncer"
	"github.com/sgostarter/libcomponents/syncer/impls/onlyrecord"
	"github.com/sgostarter/libcomponents/syncer/syncer_test"
	"github.com/sgostarter/libeasygo/pathutils"
	"github.com/stretchr/testify/assert"
)

func TestSyncer(t *testing.T) {
	_ = os.RemoveAll(utRoot)
	_ = pathutils.MustDirExists(utRoot)

	s := syncer.NewSyncer(context.Background(), nil, onlyrecord.NewMFStorage(utRoot, l.NewConsoleLoggerWrapper()), 3, l.NewConsoleLoggerWrapper())

	err := s.AppendAddRecordLog("1", []byte("1"))
	assert.Nil(t, err)

	err = s.AppendAddRecordLog("2", []byte("2"))
	assert.Nil(t, err)

	err = s.AppendAddRecordLog("3", []byte("3"))
	assert.Nil(t, err)

	err = s.AppendAddRecordLog("4", []byte("4"))
	assert.Nil(t, err)
}

func TestSyncer2(t *testing.T) {
	_ = os.RemoveAll(utRoot)
	_ = pathutils.MustDirExists(utRoot)

	s := syncer.NewSyncer(context.Background(), nil, onlyrecord.NewMFStorage(utRoot, l.NewConsoleLoggerWrapper()), 3, l.NewConsoleLoggerWrapper())

	c1 := syncert.NewUTClient(t, s)

	record1, ok := c1.AddRecord(syncert.RecordData{
		Amount: 100,
		At:     time.Now(),
		Remark: "100",
	})
	assert.True(t, ok)
	t.Log(record1)

	record2, ok := c1.AddRecord(syncert.RecordData{
		Amount: 200,
		At:     time.Now(),
		Remark: "200",
	})
	assert.True(t, ok)
	t.Log(record2)

	c1.UploadChanges()

	err := c1.SyncFromServer()
	assert.Nil(t, err)

	c2 := syncert.NewUTClient(t, s)

	err = c2.SyncFromServer()
	assert.Nil(t, err)

	assert.True(t, c1.Equal(c2))

	//
	//
	//

	record3, ok := c1.AddRecord(syncert.RecordData{
		Amount: 300,
		At:     time.Now(),
		Remark: "300",
	})
	assert.True(t, ok)
	t.Log(record3)

	ok = c1.ModifyRecord(record1, syncert.RecordData{
		Amount: 101,
		At:     time.Now(),
		Remark: "101",
	})
	assert.True(t, ok)

	c1.UploadChanges()

	ok = c1.ModifyRecord(record1, syncert.RecordData{
		Amount: 102,
		At:     time.Now(),
		Remark: "102",
	})
	assert.False(t, ok)

	err = c1.SyncFromServer()
	assert.Nil(t, err)

	err = c2.SyncFromServer()
	assert.Nil(t, err)

	assert.True(t, c1.Equal(c2))
	c1.Dump()

	//
	//
	//

	ok = c1.ModifyRecord(record1, syncert.RecordData{
		Amount: 11111,
		Remark: "11111",
	})
	assert.True(t, ok)

	ok = c2.DelRecord(record1)
	assert.True(t, ok)

	c1.UploadChanges()

	err = c1.SyncFromServer()
	assert.Nil(t, err)

	c2.UploadChanges()
	err = c2.SyncFromServer()
	assert.Nil(t, err)

	assert.True(t, c1.Equal(c2))
	c1.Dump()

	time.Sleep(time.Second * 5)

	// 200, 300, 11111
	logs, err := s.GetAllLogs("")
	assert.Nil(t, err)
	if len(logs) == 6 {
		assert.Equal(t, 6, len(logs))

		logs, err = s.GetAllLogs(logs[0].SeqID)
		assert.Nil(t, err)
		assert.Equal(t, 5, len(logs))

		logs, err = s.GetAllLogs(logs[0].SeqID)
		assert.Nil(t, err)
		assert.Equal(t, 4, len(logs))

		logs, err = s.GetAllLogs(logs[1].SeqID)
		assert.Nil(t, err)
		assert.Equal(t, 2, len(logs))

		logs, err = s.GetAllLogs(logs[1].SeqID)
		assert.Nil(t, err)
		assert.Equal(t, 0, len(logs))
	} else {
		assert.Equal(t, 4, len(logs))

		logs, err = s.GetAllLogs(logs[0].SeqID)
		assert.Nil(t, err)
		assert.Equal(t, 3, len(logs))

		logs, err = s.GetAllLogs(logs[0].SeqID)
		assert.Nil(t, err)
		assert.Equal(t, 2, len(logs))

		logs, err = s.GetAllLogs(logs[1].SeqID)
		assert.Nil(t, err)
		assert.Equal(t, 0, len(logs))
	}

}

func TestSyncer3(t *testing.T) {
	_ = os.RemoveAll(utRoot)
	_ = pathutils.MustDirExists(utRoot)

	s := syncer.NewSyncer(context.Background(), nil, onlyrecord.NewMFStorage(utRoot, l.NewConsoleLoggerWrapper()), 3, l.NewConsoleLoggerWrapper())

	c1 := syncert.NewUTClient(t, s)

	record1, ok := c1.AddRecord(syncert.RecordData{
		Amount: 100,
		At:     time.Now(),
		Remark: "100",
	})
	assert.True(t, ok)
	t.Log(record1)

	record2, ok := c1.AddRecord(syncert.RecordData{
		Amount: 200,
		At:     time.Now(),
		Remark: "200",
	})
	assert.True(t, ok)
	t.Log(record2)

	_, ok = c1.AddRecord(syncert.RecordData{
		Amount: 300,
		At:     time.Now(),
		Remark: "300",
	})
	assert.True(t, ok)

	_, ok = c1.AddRecord(syncert.RecordData{
		Amount: 400,
		At:     time.Now(),
		Remark: "400",
	})
	assert.True(t, ok)

	c1.UploadChanges()

	err := c1.SyncFromServer()
	assert.Nil(t, err)

	c2 := syncert.NewUTClient(t, s)

	err = c2.SyncFromServer()
	assert.Nil(t, err)

	assert.True(t, c1.Equal(c2))

}
