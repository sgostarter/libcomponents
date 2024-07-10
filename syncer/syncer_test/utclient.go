// nolint
package syncert

import (
	"encoding/json"
	"reflect"
	"slices"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sgostarter/libcomponents/syncer"
	"github.com/stretchr/testify/assert"
)

type RecordData struct {
	Amount int
	At     time.Time
	Remark string
}

func (rd RecordData) ToJSON() []byte {
	d, _ := json.Marshal(&rd)

	return d
}

func RecordDataFromJSON(d []byte) RecordData {
	var rd RecordData

	_ = json.Unmarshal(d, &rd)

	return rd
}

type RecordRow struct {
	ID         string
	Version    string
	UpdateFlag int // 0: not sync; 1: sync to server; 2: sync with server
	Deleted    bool
	RecordData
}

type UTClient struct {
	t *testing.T
	s syncer.Syncer

	seqID string
	rows  []*RecordRow
}

func NewUTClient(t *testing.T, s syncer.Syncer) *UTClient {
	return &UTClient{
		t:     t,
		s:     s,
		seqID: "",
		rows:  make([]*RecordRow, 0, 1),
	}
}

func (cli *UTClient) AddRecord(data RecordData) (string, bool) {
	id := uuid.NewString()

	cli.rows = append(cli.rows, &RecordRow{
		ID:         id,
		UpdateFlag: 0,
		RecordData: data,
	})

	return id, true
}

func (cli *UTClient) DelRecord(id string) bool {
	for idx := 0; idx < len(cli.rows); idx++ {
		if cli.rows[idx].ID == id {
			if cli.rows[idx].Deleted {
				cli.t.Logf("del %s record failed, already deleted it\n", id)

				return false
			}

			if cli.rows[idx].UpdateFlag == 1 { // changing
				cli.t.Logf("del %s record failed, chaning\n", id)

				return false
			}

			if cli.rows[idx].Version == "" {
				cli.t.Logf("del %s record success, only local, simple delete it\n", id)

				cli.rows = slices.Delete(cli.rows, idx, idx+1)

				return true
			}

			cli.rows[idx].Deleted = true
			cli.rows[idx].UpdateFlag = 0

			return true
		}
	}

	cli.t.Logf("del %s record failed, no record\n", id)

	return false
}

func (cli *UTClient) ModifyRecord(id string, data RecordData) bool {
	for idx := 0; idx < len(cli.rows); idx++ {
		if cli.rows[idx].ID == id {
			if cli.rows[idx].Deleted {
				cli.t.Logf("change %s record failed, already deleted it\n", id)

				return false
			}

			if cli.rows[idx].UpdateFlag == 1 { // changing
				cli.t.Logf("change %s record failed, chaning\n", id)

				return false
			}

			cli.rows[idx].UpdateFlag = 0
			cli.rows[idx].RecordData = data

			return true
		}
	}

	cli.t.Logf("change %s record failed, no record\n", id)

	return false
}

func (cli *UTClient) UploadChanges() {
	for idx := 0; idx < len(cli.rows); idx++ {
		row := cli.rows[idx]

		if row.UpdateFlag != 0 {
			continue
		}

		if row.Deleted {
			assert.Nil(cli.t, cli.s.AppendDelRecordLog(row.ID, row.Version))
		} else if row.Version == "" {
			assert.Nil(cli.t, cli.s.AppendAddRecordLog(row.ID, row.RecordData.ToJSON()))
		} else {
			assert.Nil(cli.t, cli.s.AppendChangeRecordLog(row.ID, row.Version, row.RecordData.ToJSON()))
		}

		row.UpdateFlag = 1
	}
}

func (cli *UTClient) SyncFromServer() (err error) {
	logs, err := cli.s.GetAllLogs(cli.seqID)
	if err != nil {
		cli.t.Logf("get all logs from %s failed: %v\n", cli.seqID, err)

		return
	}

	for _, log := range logs {
		if log.PluginID == "" {
			switch log.OpType {
			case syncer.OpTypeAdd:
				cli.applyAddRecord(log.RecordID, log.Ds, log.NewVersionID)
			case syncer.OpTypeDel:
				cli.applyDelRecord(log.RecordID, log.VersionID)
			case syncer.OpTypeChange:
				cli.applyChangeRecord(log.RecordID, log.VersionID, log.Ds, log.NewVersionID)
			case syncer.OpTypeSnapshot:
				cli.applySnapshot(log.Ds)
			}

			cli.seqID = syncer.SeqIDN2S(log.SeqID)
		} else {
			cli.t.Log("invalid plugin id", log.PluginID)
		}
	}

	return
}

func (cli *UTClient) applySnapshot(ds []byte) bool {
	var d syncer.SnapshotData

	err := json.Unmarshal(ds, &d)
	if err != nil {
		cli.t.Log("unmarshal snapshot data failed:", err)

		return false
	}

	cli.rows = make([]*RecordRow, 0, len(d.Records))
	for _, record := range d.Records {
		var recordData RecordData

		err = json.Unmarshal(record.Data, &recordData)
		if err != nil {
			cli.t.Log("unmarshal record data")
		}

		cli.rows = append(cli.rows, &RecordRow{
			ID:         record.ID,
			Version:    record.Version,
			UpdateFlag: int(record.UpdateFlag),
			Deleted:    record.Deleted,
			RecordData: recordData,
		})

		cli.seqID = record.ID
	}

	return true
}

func (cli *UTClient) applyChangeRecord(id string, versionID string, data []byte, newVersionID string) bool {
	for idx := 0; idx < len(cli.rows); idx++ {
		row := cli.rows[idx]

		if row.ID != id {
			continue
		}

		if row.UpdateFlag == 2 && row.Version != versionID {
			cli.t.Log("apply change record failed, invalid version id")

			return false
		}

		cli.rows[idx] = &RecordRow{
			ID:         id,
			Version:    newVersionID,
			UpdateFlag: 2,
			Deleted:    false,
			RecordData: RecordDataFromJSON(data),
		}

		return true
	}

	cli.t.Log("apply change record failed, no record")

	return false
}

func (cli *UTClient) applyDelRecord(id string, versionID string) bool {
	for idx := 0; idx < len(cli.rows); idx++ {
		row := cli.rows[idx]

		if row.ID != id {
			continue
		}

		if row.UpdateFlag == 2 && row.Version != versionID {
			cli.t.Log("apply del record failed, invalid version id")

			return false
		}

		cli.rows = slices.Delete(cli.rows, idx, idx+1)

		return true
	}

	cli.t.Log("apply del record failed, no record")

	return false
}

func (cli *UTClient) applyAddRecord(id string, data []byte, newVersionID string) bool {
	for idx := 0; idx < len(cli.rows); idx++ {
		row := cli.rows[idx]

		if row.ID != id {
			continue
		}

		if row.UpdateFlag == 2 {
			cli.t.Log("apply add record failed: update flag is 2")

			return false
		}

		aRow := RecordRow{
			ID:         id,
			Version:    newVersionID,
			UpdateFlag: 2,
			Deleted:    false,
			RecordData: RecordDataFromJSON(data),
		}

		cli.rows[idx] = &aRow

		return true
	}

	cli.rows = append(cli.rows, &RecordRow{
		ID:         id,
		Version:    newVersionID,
		UpdateFlag: 2,
		Deleted:    false,
		RecordData: RecordDataFromJSON(data)})

	return false
}

func (cli *UTClient) Equal(cli2 *UTClient) bool {
	return reflect.DeepEqual(cli.rows, cli2.rows)
}

func (cli *UTClient) Dump() {
	d, err := json.MarshalIndent(cli.rows, "", "  ")
	assert.Nil(cli.t, err)
	cli.t.Log(string(d))
}
