package curve_hashrate

import (
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/sgostarter/libcomponents/curve"
)

type ftSupport struct {
}

func (spt ftSupport) GetKey4PoolHashrate(poolID int64) string {
	return fmt.Sprintf("H_%d", poolID)
}

func (spt ftSupport) GetKey4CoinHashrate(poolID int64) string {
	return fmt.Sprintf("C_%d", poolID+1000)
}

func (spt ftSupport) GetKey4All() string {
	return "all"
}

func (spt ftSupport) GetKeys() []string {
	return []string{
		"H_1", "H_2", "C_1001", "all",
	}
}

func Test1(t *testing.T) {
	spt := &ftSupport{}
	c := curve.NewCurve[int64, Point](time.Second*2, 20, []int{1, 5, 10},
		curve.NewCommonStorage[Point]("./tmp-hashrate/"),
		&System{spt: spt}, nil)

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

	ftDump(c, spt)

	time.Sleep(time.Second * 3)

	ftDump(c, spt)

	time.Sleep(time.Second * 6)

	ftDump(c, spt)
}

func ftDump(c *curve.Curve[int64, Point], spt Supporter) {
	fmt.Println("------------------------------------------")
	tss1, ps1 := c.GetCurves(1, spt.GetKey4PoolHashrate(1), 30)
	tss2, ps2 := c.GetCurves(1, spt.GetKey4All(), 30)
	tss12, ps12 := c.GetCurves(5, spt.GetKey4PoolHashrate(1), 30)

	var ss strings.Builder

	ss.WriteString("1 - 1\n")
	for idx, at := range tss1 {
		ss.WriteString(fmt.Sprintf("  %s %v\n", time.Unix(at, 0).Format("15:04:05"), ps1[idx]))
	}
	ss.WriteString("\n")
	ss.WriteString("1 - all\n")
	for idx, at := range tss2 {
		ss.WriteString(fmt.Sprintf("  %s %v\n", time.Unix(at, 0).Format("15:04:05"), ps2[idx]))
	}
	ss.WriteString("\n")
	ss.WriteString("5 - 1\n")
	for idx, at := range tss12 {
		ss.WriteString(fmt.Sprintf("  %s %v\n", time.Unix(at, 0).Format("15:04:05"), ps12[idx]))
	}
	ss.WriteString("\n")
	fmt.Println(ss.String())
}
