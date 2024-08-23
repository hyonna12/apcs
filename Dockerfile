FROM golang:1.22-alpine

# TimeZone 설정
ENV TZ=Asia/Seoul

# 필요한 시스템 패키지 설치
RUN apk add --no-cache tzdata

# 작업 디렉터리 설정
WORKDIR /app

# go.mod와 go.sum 파일 복사 (의존성 다운로드를 위해)
COPY go.mod go.sum ./

# 의존성 다운로드
RUN go mod download

# 소스 코드 복사
COPY . .

# 프로그램 빌드
RUN go build -o main .

# 서버 포트 설정 (애플리케이션에서 사용하는 포트에 맞게 조정)
EXPOSE 3000

# 프로그램 실행
CMD go run .

