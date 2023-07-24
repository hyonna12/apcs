package model

import "time"

type Owner struct {
	OwnerId   int
	OwnerName string
	PhoneNum  string
	Address   string
	CDatetime time.Time
	UDatetime time.Time
}
