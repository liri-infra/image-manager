FROM golang:latest AS builder
ADD . /go/src/github.com/liri-infra/image-manager
WORKDIR /go/src/github.com/liri-infra/image-manager
RUN GO111MODULE=on GOARCH=amd64 CGO_ENABLED=0 GOOS=linux go build -o /image-manager

FROM alpine
MAINTAINER Pier Luigi Fiorini <pierluigi.fiorini@gmail.com>
RUN apk --no-cache add ca-certificates
RUN mkdir -p /etc/liri
WORKDIR /app
COPY --from=builder /image-manager /app/
EXPOSE 8080
CMD ["/app/image-manager", "/etc/liri/image-manager.ini"]
