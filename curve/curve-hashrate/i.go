package curve_hashrate

type Supporter interface {
	GetKey4PoolHashrate(poolID int64) string
	GetKey4CoinHashrate(poolID int64) string
	GetKey4All() string

	GetKeys() []string

	IsCsAccount(account string) bool
	IsBuildInCsAccount(account string) bool
}
