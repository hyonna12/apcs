package plc

type Event struct{}

func (e *Event) DeviceCrashEvent() {}

func (e *Event) FireDetectionEvent() {}

func (e *Event) ImpactDetectionEvent() {}
