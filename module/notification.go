package module

import log "github.com/sirupsen/logrus"

type Notification struct{}

func (n *Notification) PushNotification(msg string) {
	// 알림
	log.Info(msg)
}
