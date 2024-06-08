package bookeepingtest

import (
	"os"
	"testing"

	"github.com/sgostarter/libeasygo/pathutils"
)

const (
	utRoot = "ut-data"
)

func TestMain(m *testing.M) {
	_ = os.RemoveAll(utRoot)
	_ = pathutils.MustDirExists(utRoot)

	code := m.Run()

	//_ = os.RemoveAll(utRoot)

	os.Exit(code)
}
