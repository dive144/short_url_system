#第一阶段 构建阶段 起了一个临时的容器 
FROM golang:1.26-alpine AS builder

ENV GOPROXY=https://goproxy.cn,direct

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-s -w" -o app main.go


#第二阶段 运行阶段
FROM alpine:3.20
RUN apk --no-cache add ca-certificates tzdata

ENV TZ=Asia/Shanghai
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone


WORKDIR /app

COPY --from=builder --chown=777 /app/app .
COPY --from=builder --chown=777 /app/config ./config
RUN chmod +x app
EXPOSE 9090
CMD ["./app"]

