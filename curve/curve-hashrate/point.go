package curve_hashrate

type Point struct {
	MinerV     int64  `yaml:"hr"`
	CsV        int64  `yaml:"chr"`
	BuildInCsV int64  `yaml:"buildInCsV,omitempty"`
	Event      string `yaml:"event,omitempty"`
}
