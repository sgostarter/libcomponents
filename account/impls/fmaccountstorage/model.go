package fmaccountstorage

import (
	"github.com/sgostarter/libcomponents/account"
)

type AccountInfo struct {
	ID             uint64
	AccountName    string
	HashedPassword string
	CreateAt       int64

	Cfg *account.AdvanceConfig

	Data []byte `json:"data,omitempty" yaml:"data,omitempty"`
}
