package curve

import (
	"math/big"
	"path"
	"sync/atomic"

	"github.com/sgostarter/i/stg"
	"github.com/sgostarter/libeasygo/stg/fs/rawfs"
	"gopkg.in/yaml.v3"
)

type Int64AVGData[D int64] struct {
	sum   *big.Int
	count int
}

func GenInt64AVGData[D int64](n int64) Int64AVGData[D] {
	return Int64AVGData[D]{
		sum:   big.NewInt(n),
		count: 1,
	}
}

func (o Int64AVGData[D]) Combine(d D) ImmutableData[D] {
	z := new(big.Int).Set(o.sum)

	return Int64AVGData[D]{
		sum:   z.Add(z, big.NewInt(int64(d))),
		count: o.count + 1,
	}
}

func (o Int64AVGData[D]) Clone() ImmutableData[D] {
	return Int64AVGData[D]{
		sum:   new(big.Int).Set(o.sum),
		count: o.count,
	}
}

func (o Int64AVGData[D]) Calc() D {
	z := new(big.Int).Set(o.sum)

	return D(z.Div(z, big.NewInt(int64(o.count))).Int64())
}

//
//
//

type Int64MaxData[D int64] struct {
	v int64
}

func GenInt64MaxData[D int64](n int64) Int64MaxData[D] {
	return Int64MaxData[D]{
		v: n,
	}
}

func (o Int64MaxData[D]) Combine(d D) ImmutableData[D] {
	if o.v > int64(d) {
		return Int64MaxData[D]{
			v: o.v,
		}
	}

	return Int64MaxData[D]{
		v: int64(d),
	}
}

func (o Int64MaxData[D]) Clone() ImmutableData[D] {
	return Int64MaxData[D]{
		v: o.v,
	}
}

func (o Int64MaxData[D]) Calc() D {
	return D(o.v)
}

//
//
//

type Float64MaxData[D float64] struct {
	v float64
}

func GenFloat64MaxData[D float64](n float64) Float64MaxData[D] {
	return Float64MaxData[D]{
		v: n,
	}
}

func (o Float64MaxData[D]) Combine(d D) ImmutableData[D] {
	if o.v > float64(d) {
		return Float64MaxData[D]{
			v: o.v,
		}
	}

	return Float64MaxData[D]{
		v: float64(d),
	}
}

func (o Float64MaxData[D]) Clone() ImmutableData[D] {
	return Float64MaxData[D]{
		v: o.v,
	}
}

func (o Float64MaxData[D]) Calc() D {
	return D(o.v)
}

//
//
//

func NewCommonStorage[POINT any](root string) *CommStorage[POINT] {
	return NewCommonStorageEx[POINT](root, nil)
}

func NewCommonStorageEx[POINT any](root string, storage stg.FileStorage) *CommStorage[POINT] {
	if storage == nil {
		storage = rawfs.NewFSStorage("")
	}

	return &CommStorage[POINT]{
		root:    root,
		storage: storage,
	}
}

type CommStorage[POINT any] struct {
	root    string
	storage stg.FileStorage
	noSave  atomic.Bool
}

func (stg *CommStorage[POINT]) fileNameByKey(key string) string {
	return path.Join(stg.root, key)
}

func (stg *CommStorage[POINT]) Load(key string) (ds []*PointWithTimestamp[POINT], err error) {
	d, err := stg.storage.ReadFile(stg.fileNameByKey(key))
	if err != nil {
		return
	}

	err = yaml.Unmarshal(d, &ds)

	return
}

func (stg *CommStorage[POINT]) Save(key string, ds []*PointWithTimestamp[POINT]) (err error) {
	if stg.noSave.Load() {
		return
	}

	d, err := yaml.Marshal(ds)
	if err != nil {
		return
	}

	err = stg.storage.WriteFile(stg.fileNameByKey(key), d)

	return
}

func (stg *CommStorage[POINT]) SetSaveFlag(save bool) {
	stg.noSave.Store(!save)
}

func (stg *CommStorage[POINT]) GetSaveFlag() bool {
	return !stg.noSave.Load()
}
