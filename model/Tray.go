package model

import "time"

type Tray struct {
	TrayId       int
	TrayOccupied bool
	ItemId       int
	CDatetime    time.Time
	UDatetime    time.Time
}
