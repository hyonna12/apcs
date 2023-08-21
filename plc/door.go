package plc

type DoorType string
type DoorOperation string

const (
	DoorTypeFront DoorType = "DoorTypeFront"
	DoorTypeBack  DoorType = "DoorTypeBack"

	DoorOperationOpen  DoorOperation = "DoorOperationOpen"
	DoorOperationClose DoorOperation = "DoorOperationClose"
)

type door struct {
}
