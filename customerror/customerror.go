package customerror

type CustomError string

// error 인터페이스의 Error() 메서드를 오버라이드하여 CustomError가 error 타입으로 인식되도록 함
func (ce CustomError) Error() string {
	return string(ce)
}

// 커스텀 에러 이름 및 메시지 정의
// 에러 이름 앞에는 'Err' 표시
const (
	ErrNoEventHandlerFound         CustomError = "No event handler found to handle the dispatched event: "
	ErrMessageResponseTimeout      CustomError = "No message response received during timeout limit"
	ErrRobotJobFail                CustomError = "Robot job failed"
	ErrRobotJobRequestTimeout      CustomError = "No robot job result response during timeout limit"
	ErrRobotJobDistributionTimeout CustomError = "Job is not distributed to any robot during the timeout limit."
)
