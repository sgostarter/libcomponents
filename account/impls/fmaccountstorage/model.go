package fmaccountstorage

type AccountInfo struct {
	ID             uint64
	AccountName    string
	HashedPassword string
	CreateAt       int64
}
