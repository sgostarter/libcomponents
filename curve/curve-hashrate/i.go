package curve_hashrate

type Supporter interface {
	HRSGetKey4Pool(poolID int64) string
	HRSGetKey4Coin(poolID int64) string
	HRSGetKey4All() string

	HRSGetLoadKeys() []string

	HRSIsCsAccount(account string) bool
	HRSIsBuildInCsAccount(account string) bool
}
