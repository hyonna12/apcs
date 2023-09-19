package trayBuffer

import (
	"bytes"
	"container/list"
	"encoding/json"
	"net/http"

	log "github.com/sirupsen/logrus"
)

type BufferOperation string

const (
	BufferOperationUp   BufferOperation = "BufferOperationUp"
	BufferOperationDown BufferOperation = "BufferOperationDown"
)

var (
	Buffer *TrayBuffer
)

type TrayBufferRequest struct {
	BufferOperation BufferOperation `json:"BufferOperation"`
}

// 빈트레이 id를 담기 위한 스택
type TrayBuffer struct {
	ids *list.List
}

// SetUpTrayBuffer
//
// 트레이 버퍼 조작.
//
// - trayBuffer.BufferOperation: 조작 명령
func SetUpTrayBuffer(BufferOperation BufferOperation) error {
	log.Infof("[PLC_TrayBuffer] 트레이 버퍼 조작: %v", BufferOperation)
	// PLC 트레이 버퍼 조작
	data := TrayBufferRequest{BufferOperation: BufferOperation}
	pbytes, _ := json.Marshal(data)
	buff := bytes.NewBuffer(pbytes)
	_, err := http.Post("http://localhost:8000/setup/buffer", "application/json", buff)
	if err != nil {
		return err
	}
	return nil
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
	back := t.ids.Back().Value
	if back == nil {
		return nil
	}
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

	back := t.ids.Back()
	list = append(list, back.Value)

	prev := back.Prev()
	for prev != nil {
		list = append(list, prev.Value)
		prev = prev.Prev()
	}
	log.Debugf("buffer tray : %v", list)

	return list
}

func (t *TrayBuffer) IsEmpty() bool {
	return t.ids.Len() == 0
}

// 트레이의 개수 count 해주는 함수 추가 - db갱신하기 전 조회
func (t *TrayBuffer) Count() int {
	return t.ids.Len()
}
