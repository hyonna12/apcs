package plc

import "fmt"

type GatePlc struct{}

// 문을 제어하는 요청
func (g *GatePlc) SetUpDoor(gate, operate string) {
	// 파라미터 - frontgate/backgate, open/close
	fmt.Println(gate, operate)

}
