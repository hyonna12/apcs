package trayBuffer

import (
	"apcs_refactored/config"
	"apcs_refactored/model"
	"apcs_refactored/plc/conn"
	"container/list"
	"time"

	log "github.com/sirupsen/logrus"
)

type BufferOperation string

const (
	BufferOperationUp   BufferOperation = "BufferOperationUp"
	BufferOperationDown BufferOperation = "BufferOperationDown"
)

var (
	Buffer         *TrayBuffer
	simulatorDelay time.Duration
)

// 빈트레이 id를 담기 위한 스택
type TrayBuffer struct {
	ids *list.List
}

// SetUpTrayBuffer
//
// 트레이 버퍼 조작.
//
// - trayBuffer.BufferOperation: 조작 명령
func SetUpTrayBuffer(BufferOperation BufferOperation, commandId string) error {
	log.Infof("[PLC_Buffer] 트레이 버퍼 조작: %v", BufferOperation)

	err := conn.SendTrayBufferOperation(string(BufferOperation), commandId)
	if err != nil {
		log.Errorf("[PLC_Buffer] 버퍼 조작 실패: %v", err)
		return err
	}
	// TODO - temp - 시뮬레이터
	// time.Sleep(simulatorDelay * 500 * time.Millisecond)
	return nil
}

// 트레이 버퍼 초기 설정
func InitTrayBuffer() {
	// 트레이 버퍼 스택 생성
	Buffer := NewTrayBuffer()
	// 초기 버퍼 빈트레이 id 값
	list := make([]int, 0)
	tray := config.Config.Plc.TrayBuffer.Init
	list = append(list, tray...)

	for _, i := range list { // config.Config.Plc.TrayBuffer.Optimum
		num := int64(i)
		Buffer.Push(num)
	}
	count := Buffer.Count()

	model.InsertBufferState(count)
	Buffer.Get()
}

// 트레이 버퍼 스택 생성
func NewTrayBuffer() *TrayBuffer {
	log.Debug("TrayBuffer created")
	Buffer = &TrayBuffer{list.New()}
	return Buffer
}

// stack 에 값 추가
func (t *TrayBuffer) Push(id interface{}) {
	t.ids.PushBack(id)
}

// 맨 위의 값 반환
func (t *TrayBuffer) Peek() interface{} {
	id := t.ids.Back()
	if id == nil {
		return nil
	}
	back := id.Value

	return back
}

// 맨 위의 값 삭제하고 반환
func (t *TrayBuffer) Pop() interface{} {
	back := t.ids.Back()

	if back == nil {
		return nil
	}

	return t.ids.Remove(back)
}

// 트레이 버퍼 값 가져오기
func (t *TrayBuffer) Get() interface{} {
	list := []any{}
	if t.Count() >= 1 {
		back := t.ids.Back()
		list = append(list, back.Value)
		prev := back.Prev()
		for prev != nil {
			list = append(list, prev.Value)
			prev = prev.Prev()
		}
	}
	log.Debugf("tray buffer : %v", list)

	return list
}

func (t *TrayBuffer) IsEmpty() bool {
	return t.ids.Len() == 0
}

// 트레이의 개수 count 해주는 함수 추가 - db갱신하기 전 조회
func (t *TrayBuffer) Count() int {
	return t.ids.Len()
}
