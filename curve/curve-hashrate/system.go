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
		return fnMustSD4Key(sys.spt.GetKey4PoolHashrate(poolID))
	}

	fnMustSD4Coin := func(poolID int64) *Point {
		return fnMustSD4Key(sys.spt.GetKey4CoinHashrate(poolID))
	}

	fnMustSD4All := func() *Point {
		return fnMustSD4Key(sys.spt.GetKey4All())
	}

	for key, d := range m {
		// proxy-id:account[x,x.cs,x.cs_buildin]
		ps := strings.Split(cast.ToString(key), ":")
		if len(ps) != 2 {
			continue
		}

		proxyID, err := strconv.ParseInt(ps[0], 10, 64)
		if err != nil {
			panic("invalidProxyID")
		}

		if strings.HasSuffix(ps[1], ".cs") {
			fnMustSD4All().CsV += d
			fnMustSD4Coin(proxyID).CsV += d
			fnMustSD4Pool(proxyID).CsV += d
		} else if strings.HasSuffix(ps[1], ".cs_buildin") {
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
	return sys.spt.GetKeys()
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
