package curve

import "fmt"

type ProtocolFamily int

const (
	ProtocolFamilyTypeMin ProtocolFamily = 0
	ProtocolFamilyTypeETH                = ProtocolFamilyTypeMin
	ProtocolFamilyTypeETC ProtocolFamily = 1
)

func poolHashrateKey(proxyID int64) string {
	return fmt.Sprintf("%d", proxyID)
}

func coinTypeHashrateKey(t ProtocolFamily) string {
	return fmt.Sprintf("C_%d", t)
}

func allHashrateKey() string {
	return "0"
}
