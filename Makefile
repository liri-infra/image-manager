# SPDX-FileCopyrightText: 2019 Pier Luigi Fiorini <pierluigi.fiorini@gmail.com>
#
# SPDX-License-Identifier: CC0-1.0

DOCKER_IMAGE=liriorg/image-manager
DOCKER_VERSION=latest

all:
	@go build -v

clean:
	@rm -f image-manager

format:
	@gofmt -w .

test:
	@go test -v -cover ./...

push:
	@docker build -t $(DOCKER_IMAGE) .
	@docker push $(DOCKER_IMAGE):$(DOCKER_VERSION)
