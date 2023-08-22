package messenger

import (
	"apcs_refactored/config"
	"apcs_refactored/customerror"
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"time"
)

// 메시지 노드 이름
const (
	// NodeEventServer - 메시지 노드 이름
	NodeEventServer = NodeName("event_server")
	// NodePlcClient - 메시지 노드 이름
	NodePlcClient = NodeName("plc_server")
	// NodeWebsocketServer - 메시지 노드 이름
	NodeWebsocketServer = NodeName("websocket_server")
)

// 메시지 리프 이름
const (
	// LeafKiosk - Target 또는 Sender에 들어가는 메시지 주체
	LeafKiosk = LeafName("kiosk")
	// LeafEvent - Target 또는 Sender에 들어가는 메시지 주체
	LeafEvent = LeafName("event")
	// LeafPlc - Target 또는 Sender에 들어가는 메시지 주체
	LeafPlc = LeafName("plc")
)

type NodeName string

// LeafName - 메시지가 최종적으로 수신되는 말단 노드
type LeafName string

// Node - 메신저와 직접 연결된 노드로, Node 이면서 LeafName 일 수도 있음
type Node struct {
	Name   NodeName
	MsgHub *MsgHub
	// Listen - 허브에서 보내는 메시지를 듣는 채널
	Listen chan MessageWrapper
	// Waiting - 메시지 응답을 기다릴 필요가 있는 경우 임시로 청취하는 채널
	Waiting map[string]chan bool
	// Leaves - 메시지가 전달되는 끝단 노드 목록
	Leaves map[LeafName]struct{}
}

func NewNode(name NodeName) *Node {
	n := &Node{
		Name:    name,
		MsgHub:  nil,
		Listen:  make(chan MessageWrapper),
		Waiting: make(map[string]chan bool),
		Leaves:  make(map[LeafName]struct{}),
	}

	return n
}

// ListenMessages - 각 노드에서 메신저로부터 메시지를 청취하는 고루틴
//
//	isMessageDelivered: 각 메시지 노드에서 본인의 메시지 리프로 메시지가 잘 전달됐는지 여부를 반환하는 콜백 함수
func (n *Node) ListenMessages(isMessageDelivered func(m *Message) bool) {
	for {
		select {
		case messageWrapper := <-n.Listen:
			if messageWrapper.Node.Name == n.Name {
				continue
			}

			message := messageWrapper.Message
			log.Debugf("[%v -> %v] Recevied MessageWrapper. message: %v", messageWrapper.Node.Name, n.Name, message)

			// 본인 노드에 등록된 리프로 전달되는 메시지일 경우 응답 처리
			if _, isPresent := n.Leaves[message.Target]; isPresent {
				// 응답 메시지일 경우 - Waiting 맵에 등록된 채널에 true 전달
				if message.Data == "ok" {
					n.Waiting[message.Id] <- true
					continue
				}

				// 일반 메시지일 경우 -응답 메시지 전송
				if isMessageDelivered(message) {
					n.SendResponseMessage(messageWrapper)
				}
			}
		}

	}
}

// SendResponseMessage
//
// 응답 메시지 전송
func (n *Node) SendResponseMessage(mw MessageWrapper) {

	m := &Message{
		Id:        mw.Message.Id,
		Timestamp: mw.Message.Timestamp,
		Target:    mw.Message.Sender,
		Sender:    mw.Message.Target,
		Data:      "ok",
	}

	responseMw := MessageWrapper{Node: n, Message: m}

	select {
	case n.MsgHub.broadcast <- responseMw:
		log.Debugf("[%v -> %v] Responded to message. MessageId=%v", m.Sender, m.Target, mw.Message.Id)
	case <-time.After(time.Duration(config.Config.Messenger.Timeout.Spread) * time.Second):
		log.Errorf("[%v -> %v] Cannot respond to message. MessageId=%v", m.Sender, m.Target, mw.Message.Id)
	}
}

func (n *Node) SpreadMessage(message []byte) error {
	m := &Message{}
	if err := json.Unmarshal(message, m); err != nil {
		return err
	}

	mw := MessageWrapper{
		Node:    n,
		Message: m,
	}

	// Waiting 맵에 대기 채널 등록
	n.Waiting[m.Id] = make(chan bool)

	// 메시지 전파
	select {
	case n.MsgHub.broadcast <- mw:
		log.Debugf("[%v -> %v] Spread message. MessageId=%v", m.Sender, m.Target, mw.Message.Id)
	case <-time.After(time.Duration(config.Config.Messenger.Timeout.Spread) * time.Second):
		log.Errorf("[%v -> %v] Cannot spread message. MessageId=%v", m.Sender, m.Target, mw.Message.Id)
	}

	// Waiting 맵의 채널에 메시지 응답이 올 때까지 대기
	select {
	case <-n.Waiting[m.Id]:
		delete(n.Waiting, m.Id)
		return nil
	// 일정 시간 응답 없으면 ErrMessageResponseTimeout 에러 반환
	case <-time.After(time.Duration(config.Config.Messenger.Timeout.Response) * time.Second):
		log.Errorf("[%v -> %v] Message response timeout", m.Sender, m.Target)
		return customerror.ErrMessageResponseTimeout
	}
}
