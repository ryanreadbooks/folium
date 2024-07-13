FROM golang:1.21 AS builder
WORKDIR /build
COPY . .
ENV CGO_ENABLED=0 GO111MODULE=on GOPROXY="https://goproxy.cn,direct"
RUN go build -o folium cmd/main.go

FROM alpine:latest

WORKDIR /app
COPY --from=builder /build/folium .
ENV CGO_ENABLED=0
ENTRYPOINT [ "/app/folium" ]
