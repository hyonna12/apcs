package messenger

import (
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"time"
)

var (
	msgHub *MsgHub
	nodes  map[*Node]struct{}
)

// MessageWrapper messenger를 통해 전파되는 Message struct를 래핑하는 구조체.
//
// 메시지 전파 무한루프를 방지하기 위해 사용.
// messenger hub에 들어오고 나갈 때만 사용.
type MessageWrapper struct {
	Node    *Node    `json:"node"` // Node - 메시지 래퍼를 보내는 노드
	Message *Message `json:"message"`
}

// Message 각 메시지 주체가 인식하는 메시지
type Message struct {
	Id        string    `json:"id"`     // Id - 메시지 식별자 (UUID 사용)
	Target    LeafName  `json:"target"` // Target - 메시지를 전달받는 리프
	Sender    LeafName  `json:"sender"` // Sender - 메시지를 보내는 리프
	Timestamp time.Time `json:"timestamp"`
	Data      string    `json:"data"` // Data - 메시지 내용
}

type MsgHub struct {
	nodes     map[*Node]struct{}
	broadcast chan MessageWrapper
}

// NewMessage - 새로운 메시지 생성
//
//	t: target. 메시지 받는 리프
//	s: sender. 메시지를 보내는 리프
//	d: data. 메시지 내용
func NewMessage(t LeafName, s LeafName, d string) *Message {
	u, _ := uuid.NewRandom()

	m := &Message{
		Id:        u.String(),
		Timestamp: time.Now(),
		Target:    t,
		Sender:    s,
		Data:      d,
	}

	return m
}

func StartMessengerServer(n map[*Node]struct{}) {
	nodes = n
	msgHub = &MsgHub{
		nodes:     make(map[*Node]struct{}),
		broadcast: make(chan MessageWrapper),
	}

	// 메신저 노드 등록
	for node := range nodes {
		msgHub.nodes[node] = struct{}{}
		node.MsgHub = msgHub
	}

	go run()

	log.Debugf("Messenger started")
}

// run - broadcast 채널로 들어오는 메시지를 모든 노드에 전파하는 goroutine
func run() {
	for {
		select {
		case messageWrapper := <-msgHub.broadcast:
			for node := range msgHub.nodes {
				// 메시지를 보낸 노드일 경우 전달하지 않음 (무한루프 방지)
				if node != messageWrapper.Node {
					select {
					case node.Listen <- messageWrapper:
					case <-time.After(1 * time.Second):
						log.Errorf("[Messenger -> %v] Cannot send message wrapper", node.Name)
					}
				}
			}
		}
	}
}
