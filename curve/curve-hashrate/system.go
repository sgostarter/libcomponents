package curve_hashrate

import (
	"strconv"
	"strings"

	"github.com/sgostarter/libcomponents/curve"
	"github.com/spf13/cast"
)

type System struct {
	spt Supporter
}

func NewSystem(spt Supporter) *System {
	return &System{spt: spt}
}

func (sys *System) ExplainDataAt(m map[string]int64) map[string]Point {
	rm := make(map[string]*Point)

	fnMustSD4Key := func(key string) *Point {
		sd, exists := rm[key]
		if exists {
			return sd
		}

		sd = &Point{}
		rm[key] = sd

		return sd
	}

	fnMustSD4Pool := func(poolID int64) *Point {
		return fnMustSD4Key(sys.spt.HRSGetKey4Pool(poolID))
	}

	fnMustSD4Coin := func(poolID int64) *Point {
		return fnMustSD4Key(sys.spt.HRSGetKey4Coin(poolID))
	}

	fnMustSD4All := func() *Point {
		return fnMustSD4Key(sys.spt.HRSGetKey4All())
	}

	for key, d := range m {
		// proxy-id:account[x,x.cs,x.cs_buildin]
		ps := strings.SplitN(cast.ToString(key), ":", 2)
		if len(ps) != 2 {
			continue
		}

		proxyID, err := strconv.ParseInt(ps[0], 10, 64)
		if err != nil {
			panic("invalidProxyID")
		}

		if sys.spt.HRSIsCsAccount(ps[1]) {
			fnMustSD4All().CsV += d
			fnMustSD4Coin(proxyID).CsV += d
			fnMustSD4Pool(proxyID).CsV += d
		} else if sys.spt.HRSIsBuildInCsAccount(ps[1]) {
			fnMustSD4All().BuildInCsV += d
			fnMustSD4Coin(proxyID).BuildInCsV += d
			fnMustSD4Pool(proxyID).BuildInCsV += d
		} else {
			fnMustSD4All().MinerV += d
			fnMustSD4Coin(proxyID).MinerV += d
			fnMustSD4Pool(proxyID).MinerV += d
		}
	}

	retM := make(map[string]Point)

	for s, point := range rm {
		retM[s] = *point
	}

	return retM
}

func (sys *System) GetKeys() []string {
	return sys.spt.HRSGetLoadKeys()
}

func (sys *System) NewImmutableData(d int64) curve.ImmutableData[int64] {
	return curve.GenInt64AVGData[int64](d)
}

func (sys *System) NewPOINT() Point {
	return Point{}
}

func (sys *System) IsEmptyPOINT(p Point) bool {
	return p.MinerV == 0 && p.CsV == 0
}
