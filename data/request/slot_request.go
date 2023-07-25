package request

type SlotUpdateRequest struct {
	SlotEnabled bool
	SlotKeepCnt int
	TrayId      int
	ItemId      int
	Lane        int
	Floor       int
}
