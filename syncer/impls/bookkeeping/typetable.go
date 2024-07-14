package bookkeeping

import (
	"sync"

	"github.com/sgostarter/i/l"
	"github.com/sgostarter/i/stg"
	"github.com/sgostarter/libeasygo/ptl"
	"github.com/sgostarter/libeasygo/stg/mwf"
)

func NewMFTypeTable(file string, logger l.Wrapper) TypeTable {
	return NewMFTypeTableEx(file, nil, logger)
}

func NewMFTypeTableEx(file string, storage stg.FileStorage, logger l.Wrapper) TypeTable {
	if logger == nil {
		logger = l.NewNopLoggerWrapper()
	}

	logger = logger.WithFields(l.StringField(l.ClsKey, "typeTableImpl"))

	impl := &typeTableImpl{
		logger:         logger,
		validParentIDs: make(map[string]any),
	}

	impl.init(file, storage)

	return impl
}

type typeRow struct {
	Label      string
	ParentID   string
	ToParentID string
	Data       []byte
}

type typeTableImpl struct {
	logger l.Wrapper

	d              *mwf.MemWithFile[map[string]*typeRow, mwf.Serial, mwf.Lock]
	validParentIDs map[string]any
}

func (impl *typeTableImpl) BeforeLoad() {

}

func (impl *typeTableImpl) AfterLoad(vm map[string]*typeRow, err error) {
	impl.validParentIDs = make(map[string]any)

	if err != nil {
		return
	}

	for _, row := range vm {
		if row.ToParentID == "" && row.ParentID != "" {
			impl.validParentIDs[row.ParentID] = true
		}
	}
}

func (impl *typeTableImpl) BeforeSave() {

}

func (impl *typeTableImpl) AfterSave(_ map[string]*typeRow, _ error) {

}

func (impl *typeTableImpl) init(file string, storage stg.FileStorage) {
	impl.d = mwf.NewMemWithFileEx[map[string]*typeRow, mwf.Serial, mwf.Lock](make(map[string]*typeRow),
		&mwf.JSONSerial{}, &sync.RWMutex{}, file, storage, impl)
}

func (impl *typeTableImpl) Reset() {
	_ = impl.d.Change(func(_ map[string]*typeRow) (newV map[string]*typeRow, err error) {
		newV = make(map[string]*typeRow)

		return
	})

	impl.validParentIDs = make(map[string]any)
}

func (impl *typeTableImpl) TestAdd(id, label, parentID string) (ok bool, err error) {
	impl.d.Read(func(vm map[string]*typeRow) {
		_, exists := vm[id]
		if exists {
			return
		}

		if parentID != "" {
			if parentID == id {
				return
			}

			var parentRow *typeRow

			parentRow, ok = vm[parentID]
			if !ok {
				return
			}

			if parentRow.ToParentID != "" {
				return
			}

			if parentRow.ParentID != "" {
				return
			}
		}

		for _, row := range vm {
			if row.Label == label {
				return
			}
		}

		ok = true
	})

	return
}

func (impl *typeTableImpl) Add(id, label, parentID string, data []byte) error {
	return impl.d.Change(func(v map[string]*typeRow) (newV map[string]*typeRow, err error) {
		newV = v

		if newV == nil {
			newV = make(map[string]*typeRow)
		}

		_, ok := newV[id]
		if ok {
			err = ptl.NewCodeError(ptl.CodeErrExists)

			impl.logger.WithFields(l.StringField("id", id)).Error("add: id exists")

			return
		}

		if parentID != "" {
			if parentID == id {
				err = ptl.NewCodeError(ptl.CodeErrConflict)

				impl.logger.WithFields(l.StringField("id", id)).Error("add: id eq parent id")

				return
			}

			var parentRow *typeRow

			parentRow, ok = newV[parentID]
			if !ok {
				err = ptl.NewCodeError(ptl.CodeErrNotExists)

				impl.logger.WithFields(l.StringField("id", id)).Error("add: parent id not exists")

				return
			}

			if parentRow.ToParentID != "" {
				err = ptl.NewCodeError(ptl.CodeErrConflict)

				impl.logger.WithFields(l.StringField("id", id)).Error("add: parent transferred")

				return
			}

			if parentRow.ParentID != "" {
				err = ptl.NewCodeError(ptl.CodeErrConflict)

				impl.logger.WithFields(l.StringField("id", id)).Error("add: parent is a child")

				return
			}
		}

		for cID, row := range newV {
			if row.Label == label {
				err = ptl.NewCodeError(CodeErrLabelExists)

				impl.logger.WithFields(l.StringField("id", id), l.StringField("cID", cID)).
					Error("add: label dup")

				return
			}
		}

		if parentID != "" {
			impl.validParentIDs[parentID] = true
		}

		newV[id] = &typeRow{
			Label:    label,
			Data:     data,
			ParentID: parentID,
		}

		return
	})
}

func (impl *typeTableImpl) TestDel(id, toRecordID string) (ok bool, err error) {
	impl.d.Read(func(vm map[string]*typeRow) {
		if id == toRecordID {
			return
		}

		_, exists := vm[id]
		if !exists {
			return
		}

		toRow, exists := vm[toRecordID]
		if !exists {
			return
		}

		if toRow.ToParentID != "" {
			return
		}

		_, ok = impl.validParentIDs[id]
		if ok {
			return
		}

		ok = true
	})

	return
}

func (impl *typeTableImpl) Del(id, toRecordID string) error {
	return impl.d.Change(func(v map[string]*typeRow) (newV map[string]*typeRow, err error) {
		newV = v

		if newV == nil {
			newV = make(map[string]*typeRow)
		}

		if id == toRecordID {
			err = ptl.NewCodeError(ptl.CodeErrConflict)

			impl.logger.WithFields(l.StringField("id", id)).Error("del: trans to self")

			return
		}

		row, ok := newV[id]
		if !ok {
			err = ptl.NewCodeError(ptl.CodeErrNotExists)

			impl.logger.WithFields(l.StringField("id", id)).Error("del: not exists")

			return
		}

		toRow, ok := newV[toRecordID]
		if !ok {
			err = ptl.NewCodeError(ptl.CodeErrNotExists)

			impl.logger.WithFields(l.StringField("id", id)).Error("del: to not exists")

			return
		}

		if toRow.ToParentID != "" {
			err = ptl.NewCodeError(ptl.CodeErrConflict)

			impl.logger.WithFields(l.StringField("id", id)).Error("del: to has trans")

			return
		}

		_, ok = impl.validParentIDs[id]
		if ok {
			err = ptl.NewCodeError(ptl.CodeErrConflict)

			impl.logger.WithFields(l.StringField("id", id)).Error("del: has child node")

			return
		}

		if row.ParentID != "" {
			var parentOk bool

			for i, r := range newV {
				if i == id {
					continue
				}

				if r.ParentID == row.ParentID {
					parentOk = true

					break
				}
			}

			if !parentOk {
				delete(impl.validParentIDs, row.ParentID)
			}
		}

		row.ToParentID = toRecordID

		return
	})
}

func (impl *typeTableImpl) TestChange(id, label, parentID string) (ok bool, err error) {
	impl.d.Read(func(vm map[string]*typeRow) {
		_, exists := vm[id]
		if !exists {
			return
		}

		// nolint: nestif
		if parentID != "" {
			if parentID == id {
				return
			}

			_, ok = impl.validParentIDs[id]
			if ok {
				return
			}

			var parentRow *typeRow

			parentRow, ok = vm[parentID]
			if !ok {
				return
			}

			if parentRow.ParentID != "" {
				return
			}

			if parentRow.ToParentID != "" {
				return
			}
		}

		for cID, row := range vm {
			if cID == id {
				continue
			}

			if row.Label == label {
				return
			}
		}

		ok = true
	})

	return
}

// nolint: gocognit
func (impl *typeTableImpl) Change(id, label, parentID string, data []byte) error {
	return impl.d.Change(func(v map[string]*typeRow) (newV map[string]*typeRow, err error) {
		newV = v

		if newV == nil {
			newV = make(map[string]*typeRow)
		}

		tr, ok := newV[id]
		if !ok {
			err = ptl.NewCodeError(ptl.CodeErrNotExists)

			impl.logger.WithFields(l.StringField("id", id)).Error("change: no id type")

			return
		}

		var checkParentOk bool

		// nolint: nestif
		if parentID != "" {
			if parentID == id {
				err = ptl.NewCodeError(ptl.CodeErrConflict)

				impl.logger.WithFields(l.StringField("id", id)).Error("change: id eq parent id")

				return
			}

			_, ok = impl.validParentIDs[id]
			if ok {
				err = ptl.NewCodeError(ptl.CodeErrConflict)

				impl.logger.WithFields(l.StringField("id", id)).Error("change: id is parent, should not have parent id")

				return
			}

			var parentRow *typeRow

			parentRow, ok = newV[parentID]
			if !ok {
				err = ptl.NewCodeError(ptl.CodeErrConflict)

				impl.logger.WithFields(l.StringField("id", id)).Error("change: parent id not exists")

				return
			}

			if parentRow.ParentID != "" {
				err = ptl.NewCodeError(ptl.CodeErrConflict)

				impl.logger.WithFields(l.StringField("id", id)).Error("add: parent is a child")

				return
			}

			if parentRow.ToParentID != "" {
				err = ptl.NewCodeError(ptl.CodeErrConflict)

				impl.logger.WithFields(l.StringField("id", id)).Error("add: parent has trans")

				return
			}
		}

		if tr.ParentID != "" && tr.ParentID != parentID {
			checkParentOk = true
		}

		var parentOk bool

		for cID, row := range newV {
			if cID == id {
				continue
			}

			if row.Label == label {
				err = ptl.NewCodeError(CodeErrLabelExists)

				impl.logger.WithFields(l.StringField("id", id)).Error("change: label dup")

				return
			}

			if checkParentOk {
				if row.ParentID == tr.ParentID {
					parentOk = true
				}
			}
		}

		if checkParentOk && !parentOk {
			delete(impl.validParentIDs, tr.ParentID)
		}

		if parentID != "" {
			impl.validParentIDs[parentID] = true
		}

		tr.Label = label
		tr.Data = data
		tr.ParentID = parentID

		return
	})
}
