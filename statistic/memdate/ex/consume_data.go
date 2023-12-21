package ex

type LifeCostTotalData struct {
	ConsumeCount  int `json:"consume_count,omitempty"`
	ConsumeAmount int `json:"consume_amount,omitempty"`
	EarnCount     int `json:"earn_count,omitempty"`
	EarnAmount    int `json:"earn_amount,omitempty"`
}

type ListCostDataType int

const (
	ListCostDataNon ListCostDataType = iota
	ListCostDataAdd
	ListCostDataReplace
)

type LifeCostData struct {
	T             ListCostDataType
	ConsumeCount  int
	ConsumeAmount int
	EarnCount     int
	EarnAmount    int
}

type LifeCostDataTrans struct {
}

func (LifeCostDataTrans) Combine(totalD *LifeCostTotalData, d LifeCostData) *LifeCostTotalData {
	r := *totalD

	switch d.T {
	case ListCostDataAdd:
		r.ConsumeCount += d.ConsumeCount
		r.ConsumeAmount += d.ConsumeAmount
		r.EarnCount += d.EarnCount
		r.EarnAmount += d.EarnAmount
	case ListCostDataReplace:
		r.ConsumeCount = d.ConsumeCount
		r.ConsumeAmount = d.ConsumeAmount
		r.EarnCount = d.EarnCount
		r.EarnAmount = d.EarnAmount
	}

	return &r
}
