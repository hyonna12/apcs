package webserver

type KioskHandler struct{}

func NewKioskHandler() *KioskHandler {
	return &KioskHandler{}
}

func (h *KioskHandler) ChangeView(view string) error {
	return ChangeKioskView(view)
}

func (h *KioskHandler) SendEvent(event string) error {
	SendEvent(event)
	return nil
}
