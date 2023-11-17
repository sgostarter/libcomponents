package curve

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/spf13/cast"
)

type UTBizPoint struct {
	Hashrate          int64
	CsHashrate        int64
	BuildInCsHashrate int64
}

type UTBizSystem struct {
}

// nolint
func (UTBizSystem) ExplainDataAt(m map[string]int64) map[string]UTBizPoint {
	rm := make(map[string]*UTBizPoint)

	fnMustSD4Key := func(key string) *UTBizPoint {
		sd, exists := rm[key]
		if exists {
			return sd
		}

		sd = &UTBizPoint{}
		rm[key] = sd

		return sd
	}

	fnMustSD4Pool := func(poolID int64) *UTBizPoint {
		return fnMustSD4Key(poolHashrateKey(poolID))
	}

	fnMustSD4Coin := func(poolID int64) *UTBizPoint {
		return fnMustSD4Key(coinTypeHashrateKey(1)) // TODO
	}

	fnMustSD4All := func() *UTBizPoint {
		return fnMustSD4Key(allHashrateKey())
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

		/*
			key:
				all: 0
				coin: C_<coin:1,2,3>
				pool: <pool-id>
		*/

		if strings.HasSuffix(ps[1], ".cs") {
			fnMustSD4All().CsHashrate += d
			fnMustSD4Coin(proxyID).CsHashrate += d
			fnMustSD4Pool(proxyID).CsHashrate += d
		} else if strings.HasSuffix(ps[1], ".cs_buildin") {
			fnMustSD4All().BuildInCsHashrate += d
			fnMustSD4Coin(proxyID).BuildInCsHashrate += d
			fnMustSD4Pool(proxyID).BuildInCsHashrate += d
		} else {
			fnMustSD4All().Hashrate += d
			fnMustSD4Coin(proxyID).Hashrate += d
			fnMustSD4Pool(proxyID).Hashrate += d
		}
	}

	retM := make(map[string]UTBizPoint)

	for s, point := range rm {
		retM[s] = *point
	}

	return retM
}

func (UTBizSystem) GetKeys() []string {
	return []string{allHashrateKey(), poolHashrateKey(1), coinTypeHashrateKey(1)}
}

func (UTBizSystem) NewImmutableData(d int64) ImmutableData[int64] {
	return GenInt64AVGData[int64](d)
}

func (UTBizSystem) NewPOINT() UTBizPoint {
	return UTBizPoint{}
}

func (UTBizSystem) IsEmptyPOINT(p UTBizPoint) bool {
	return p.Hashrate == 0 || p.CsHashrate == 0
}

type utObserver struct {
	t *testing.T
}

// nolint
func (impl *utObserver) OnUpdate(samples []*DataUpdateSample[UTBizPoint]) {
	impl.t.Log("OnUpdate ==========>")
	for _, sample := range samples {
		impl.t.Log(sample.Speed, time.Unix(sample.At, 0).Format("15:04:05"))
		for key, data := range sample.Samples {
			impl.t.Log("  ", key, data.D)
		}
	}
	impl.t.Log("OnUpdate <==========")
}

// nolint
func TestCurve1(t *testing.T) {
	c := NewCurveEx[int64, UTBizPoint](time.Second*2, 20, []int{1, 5, 10}, NewCommonStorage[UTBizPoint]("./tmp/"),
		&UTBizSystem{}, &utObserver{t: t}, nil)

	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		defer wg.Done()

		for i := 0; i < 30; i++ {
			c.SetData("1:zjz.cs", 1)
			time.Sleep(time.Millisecond * 500)
		}
	}()

	wg.Add(1)

	go func() {
		defer wg.Done()

		for i := 0; i < 30; i++ {
			c.SetData("1:zjz", 2)
			time.Sleep(time.Millisecond * 500)
		}
	}()

	wg.Add(1)

	go func() {
		defer wg.Done()

		for i := 0; i < 30; i++ {
			c.SetData("1:zjz.cs_buildin", 3)
			time.Sleep(time.Millisecond * 500)
		}
	}()

	wg.Add(1)

	go func() {
		defer wg.Done()

		for i := 0; i < 30; i++ {
			c.SetData("2:zjc", 20)
			time.Sleep(time.Millisecond * 500)
		}
	}()

	wg.Wait()

	for idx := 0; idx < 5; idx++ {
		tss1, ps1 := c.GetCurves(1, poolHashrateKey(1), 30)
		tss2, ps2 := c.GetCurves(1, allHashrateKey(), 30)
		tss12, ps12 := c.GetCurves(5, poolHashrateKey(1), 30)

		var ss strings.Builder

		ss.WriteString("1 - 1\n")
		for idx, at := range tss1 {
			ss.WriteString(fmt.Sprintf("  %s %d\n", time.Unix(at, 0).Format("15:04:05"), ps1[idx]))
		}
		ss.WriteString("\n")
		ss.WriteString("1 - all\n")
		for idx, at := range tss2 {
			ss.WriteString(fmt.Sprintf("  %s %d\n", time.Unix(at, 0).Format("15:04:05"), ps2[idx]))
		}
		ss.WriteString("\n")
		ss.WriteString("5 - 1\n")
		for idx, at := range tss12 {
			ss.WriteString(fmt.Sprintf("  %s %d\n", time.Unix(at, 0).Format("15:04:05"), ps12[idx]))
		}
		ss.WriteString("\n")
		fmt.Println(ss.String())

		time.Sleep(time.Second * 2)
	}

}
