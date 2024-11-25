package webserver

import (
	"apcs_refactored/event/trouble"

	log "github.com/sirupsen/logrus"
)

type TroubleHandler struct{}

func (h *TroubleHandler) HandleTroubleEvent(troubleType string, details map[string]interface{}) error {
	switch troubleType {
	case trouble.NetworkDisconnected:
		if status, ok := details["status"].(string); ok && status == "disconnected" {
			log.Warn("[PLC] 네트워크 절체 감지")
			if err := ChangeKioskView("/error/trouble"); err != nil {
				log.Error("Error changing view:", err)
				return err
			}
			SendEvent("네트워크 절체")
		}
	}
	return nil
}

func NewTroubleHandler() trouble.TroubleEventHandler {
	return &TroubleHandler{}
}
