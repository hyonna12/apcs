package plc

import log "github.com/sirupsen/logrus"

type GatePlc struct{}

// 문을 제어하는 요청
func (g *GatePlc) SetUpDoor(gate, operate string) {
	// 파라미터 - frontgate/backgate, open/close
	log.Info(gate, operate)

}
