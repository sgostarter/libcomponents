// nolint
package bookkeeping

import (
	"os"
	"testing"

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

func TestTypeTable(t *testing.T) {
	_ = os.RemoveAll(utRoot)
	_ = pathutils.MustDirExists(utRoot)

	tt := NewMFTypeTableEx("tt", rawfs.NewFSStorage(utRoot), nil)

	err := tt.Add("1", "1", "", []byte("111"))
	assert.Nil(t, err)

	err = tt.Add("1", "2", "", []byte("111"))
	assert.NotNil(t, err)

	err = tt.Add("2", "1", "", []byte("111"))
	assert.NotNil(t, err)

	err = tt.Add("2", "2", "", []byte("111"))
	assert.Nil(t, err)

	err = tt.Change("2", "x2", "", []byte("x2"))
	assert.Nil(t, err)

	err = tt.Change("2", "1", "", []byte("x2"))
	assert.NotNil(t, err)

	err = tt.Del("2")
	assert.Nil(t, err)

	err = tt.Add("2", "2", "", []byte("222"))
	assert.Nil(t, err)
}

func TestTypeTable2(t *testing.T) {
	_ = os.RemoveAll(utRoot)
	_ = pathutils.MustDirExists(utRoot)

	tt := NewMFTypeTableEx("tt", rawfs.NewFSStorage(utRoot), nil)

	err := tt.Add("1", "1", "", []byte("111"))
	assert.Nil(t, err)

	err = tt.Add("2", "2", "", []byte("222"))
	assert.Nil(t, err)

	err = tt.Add("3", "3", "", []byte("333"))
	assert.Nil(t, err)

	err = tt.Add("4", "4", "1", []byte("444"))
	assert.Nil(t, err)

	err = tt.Add("5", "5", "100", []byte("555"))
	assert.NotNil(t, err)

	err = tt.Add("5", "5", "5", []byte("555"))
	assert.NotNil(t, err)

	err = tt.Change("1", "x", "2", []byte("xxx"))
	assert.NotNil(t, err)

	err = tt.Del("1")
	assert.NotNil(t, err)

	err = tt.Del("4")
	assert.Nil(t, err)

	err = tt.Del("1")
	assert.Nil(t, err)
}
