package plc

type EventPlc struct{}

func (e *EventPlc) DeviceCrashEvent() {}

func (e *EventPlc) FireDetectionEvent() {}

func (e *EventPlc) ImpactDetectionEvent() {}
