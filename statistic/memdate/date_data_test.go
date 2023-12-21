package memdate

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDateData(t *testing.T) {
	yD := NewYearData[string](2023, time.Local)

	d, err := json.MarshalIndent(yD, "", "  ")
	assert.Nil(t, err)

	t.Log(string(d))
}

func TestGetKeysForAt(t *testing.T) {
	at := time.Date(2023, 12, 21, 0, 0, 0, 0, time.Local)
	year, season, month, week, weekDay, ok := GetKeysForAt(at)
	assert.True(t, ok)
	t.Log(year, season, month, week, weekDay)
}
