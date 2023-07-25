package response

type SlotReadResponse struct {
	SlotId            int
	Lane              int
	Floor             int
	TransportDistance int
	SlotEnabled       bool
	SlotKeepCnt       int
	TrayId            int
	ItemId            int
}
