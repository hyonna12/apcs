package model

import "time"

type Slot struct {
	SlotId           int
	Lane             int
	Floor            int
	TranportDistance int
	SlotEnabled      bool
	SlotKeepCnt      int
	TrayId           int
	ItemId           int
	CheckDatetime    time.Time
	CDatetime        time.Time
	UDatetime        time.Time
}