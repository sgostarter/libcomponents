package curve

type ImmutableData[D any] interface {
	Combine(d D) ImmutableData[D]
	Calc() D
	Clone() ImmutableData[D]
}

type PointWithTimestamp[POINT any] struct {
	At int64 `yaml:"at"`
	D  POINT `yaml:",inline"`
}

type Storage[POINT any] interface {
	Load(key string) (ds []*PointWithTimestamp[POINT], err error)
	Save(key string, ds []*PointWithTimestamp[POINT]) error
}

type BizSystem[D, POINT any] interface {
	ExplainDataAt(m map[string]D) map[string]POINT

	GetKeys() []string

	NewImmutableData(d D) ImmutableData[D]
	NewPOINT() POINT
	IsEmptyPOINT(p POINT) bool
}
