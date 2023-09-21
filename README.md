# package 구조

APCS
│
└─ config
│
└─ event
│
└─ messenger --> 메시지 전파
│
└─ model
│    │
│    └─  Admin.go --> 관리자 Model
│    └─  Delivery.go --> 물품기사 Model
│    └─  Item.go --> 물품 Model
│    └─  Owner.go --> 입주민 Model
│    └─  Slot.go --> 슬롯 Model
│    └─  Tray.go --> 트레이 Model
│
└─ plc --> plc 모듈
│    │
│    └─ door --> 도어 관리
│    │   │
│    │   └─  door.go --> 도어 제어
│    │
│    └─ resource --> 택배함 자원 관리
│    │   │
│    │   └─ resource.go --> 슬롯, 테이블 점유
│    │
│    └─ robot --> 로봇 관리
│    │   │
│    │   └─  job.go --> 로봇 작업 단위 관리
│    │   └─  robot.go --> 로봇 제어
│    │
│    └─ sensor --> 센서 관리
│    │   │
│    │   └─  sensor.go --> 센서 제어
│    │
│    └─ trayBuffer --> 트레이버퍼 관리
│    │   │
│    │   └─  trayBuffer.go --> 트레이버퍼 제어
│    │
│    └─ plc.go --> plc 기능
│
└─ webserver
│    │
│    └─ static
│    │
│    └─ views --> HTML
│    │
│    └─ handler.go
│    │
│    └─ input.go --> 수납 기능
│    │
│    └─ output.go --> 불출 기능
│    │
│    └─ router.go 
│    │
│    └─ server.go 
│    │
│    └─ sort.go --> 정리 기능
│    │
│    └─ template.go 
│    │
│    └─ wsclient.go --> 웹소켓 클라이언트
│    │
│    └─ wshub.go --> 웹소켓 서버
│
└─ go.mod --> 의존하는 모듈과 버전 관리
│
└─ go.sum --> 종속성 관리를 위한 체크섬 정보 저장
│
└─ main.go
