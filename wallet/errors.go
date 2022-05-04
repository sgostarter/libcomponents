package wallet

import "errors"

var (
	ErrFailed    = errors.New("failed")
	ErrBadData   = errors.New("bad data")
	ErrNoCoins   = errors.New("no coins")
	ErrExists    = errors.New("exists")
	ErrNotExists = errors.New("not exists")
)
