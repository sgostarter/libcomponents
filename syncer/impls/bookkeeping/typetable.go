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
	Label    string
	ParentID string
	Data     []byte
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

	for id, row := range vm {
		if row.ParentID != "" {
			impl.validParentIDs[id] = true
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

			_, ok = newV[parentID]
			if !ok {
				err = ptl.NewCodeError(ptl.CodeErrNotExists)

				impl.logger.WithFields(l.StringField("id", id)).Error("add: parent id not exists")

				return
			}

			impl.validParentIDs[parentID] = true
		}

		for cID, row := range newV {
			if row.Label == label {
				err = ptl.NewCodeError(CodeErrLabelExists)

				impl.logger.WithFields(l.StringField("id", id), l.StringField("cID", cID)).
					Error("add: label dup")

				return
			}
		}

		newV[id] = &typeRow{
			Label:    label,
			Data:     data,
			ParentID: parentID,
		}

		return
	})
}

func (impl *typeTableImpl) Del(id string) error {
	return impl.d.Change(func(v map[string]*typeRow) (newV map[string]*typeRow, err error) {
		newV = v

		if newV == nil {
			newV = make(map[string]*typeRow)
		}

		row, ok := newV[id]
		if ok { // nolint: nestif
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

			_, ok = impl.validParentIDs[id]
			if ok {
				err = ptl.NewCodeError(ptl.CodeErrConflict)

				impl.logger.WithFields(l.StringField("id", id)).Error("del: has child node")

				return
			}

			delete(newV, id)
		}

		return
	})
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

			_, ok = newV[parentID]
			if !ok {
				err = ptl.NewCodeError(ptl.CodeErrConflict)

				impl.logger.WithFields(l.StringField("id", id)).Error("change: parent id not exists")

				return
			}

			impl.validParentIDs[parentID] = true
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

		tr.Label = label
		tr.Data = data
		tr.ParentID = parentID

		return
	})
}
