FROM golang:latest AS builder
RUN go get -d -v github.com/gorilla/mux \
    && go get -d -v github.com/gorilla/handlers \
    && go get -d -v gopkg.in/gcfg.v1
ADD . /go/src/github.com/liri-infra/image-manager
RUN GOARCH=amd64 CGO_ENABLED=0 GOOS=linux go build -o /image-manager github.com/liri-infra/image-manager

FROM alpine
MAINTAINER Pier Luigi Fiorini <pierluigi.fiorini@gmail.com>
RUN apk --no-cache add ca-certificates
RUN mkdir -p /etc/liri
WORKDIR /app
COPY --from=builder /image-manager /app/
EXPOSE 8080
CMD ["/app/image-manager", "/etc/liri/image-manager.ini"]
