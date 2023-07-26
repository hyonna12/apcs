package module

import "fmt"

type Notification struct{}

func (n *Notification) PushNotification(msg string) {
	// 알림
	fmt.Println(msg)
}
