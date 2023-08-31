package main

import (
	"apcs_refactored/config"
	"apcs_refactored/event"
	"apcs_refactored/messenger"
	"apcs_refactored/model"
	"apcs_refactored/plc"
	"apcs_refactored/plc/resource"
	"apcs_refactored/webserver"

	log "github.com/sirupsen/logrus"
)

func main() {

	// 설정 변수 초기화
	config.InitConfig()

	// logger 세팅
	log.SetReportCaller(true) // 파일명, 호출 라인 표시

	if config.Config.Profile == "prod" {
		// JSON 출력 - 운영 단계에서 설정
		log.SetFormatter(&log.JSONFormatter{})
	} else {
		// 로깅 텍스트 출력 - 개발 단계에서 설정
		// 무슨 이유에서인가 잘 안 먹힘
		// TODO - output stream 빼서 보기 좋게 포매팅하기
		log.SetFormatter(&log.TextFormatter{
			//ForceColors:     true,
			TimestampFormat: "15:04:05",
			//PadLevelText:    true,
			DisableLevelTruncation: false,
		})
	}

	log.Info("Control server started")

	loggingLevel := config.Config.Logging.Level
	switch loggingLevel {
	case "trace":
		log.SetLevel(log.TraceLevel)
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	}

	// DB 초기화
	model.InitDB()
	// DB connection close 지연 호출
	defer model.CloseDB()

	// 각 메시지 노드 및 노드별 리프(끝단 노드) 등록
	msgNodes := make(map[*messenger.Node]struct{})

	eventMsgNode := messenger.NewNode(messenger.NodeEventServer)
	eventMsgNode.Leaves[messenger.LeafEvent] = struct{}{}

	websocketserverMsgNode := messenger.NewNode(messenger.NodeWebsocketServer)
	websocketserverMsgNode.Leaves[messenger.LeafKiosk] = struct{}{} // TODO - node kiosk 붙인 후 주석처리 필수 (kiosk 리프 충돌)

	plcMsgNode := messenger.NewNode(messenger.NodePlcClient)
	plcMsgNode.Leaves[messenger.LeafPlc] = struct{}{}

	msgNodes[eventMsgNode] = struct{}{}
	msgNodes[websocketserverMsgNode] = struct{}{}
	msgNodes[plcMsgNode] = struct{}{}

	// 메신저 서버 시작
	messenger.StartMessengerServer(msgNodes)
	// PLC 클라이언트 시작
	plc.StartPlcClient(plcMsgNode)
	// PLC 리소스 초기화
	slots, err := model.SelectSlotList()
	if err != nil {
		log.Panicf("Failed to initialize PLC resources. %v", err)
	}
	slotIds := make([]int64, 0)
	for _, slot := range slots {
		slotIds = append(slotIds, slot.SlotId)
	}
	resource.InitResources(slotIds)
	// 이벤트 서버 시작
	event.StartEventServer(eventMsgNode)
	// 웹소켓 서버 시작
	webserver.StartWebserver(websocketserverMsgNode)
}
