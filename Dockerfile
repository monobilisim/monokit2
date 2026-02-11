FROM golang:1.26-trixie

ENV container=docker
ENV DEBIAN_FRONTEND=noninteractive
ENV CGO_ENABLED=1

WORKDIR /app

RUN apt-get update && \
    apt-get install -y --no-install-recommends \
        build-essential \
        e2fsprogs \
        systemd \
        systemd-sysv \
        dbus && \
    rm -rf /var/lib/apt/lists/*

STOPSIGNAL SIGRTMIN+3

COPY go.mod go.sum ./

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download

COPY . .

RUN mkdir -p /etc/mono && \
    cp config/* /etc/mono && \
    chmod +x scripts/collect-test-artifacts.sh

COPY scripts/docker-tests.service /etc/systemd/system/docker-tests.service
COPY scripts/exit.target /etc/systemd/system/exit.target
COPY scripts/exit-code.service /etc/systemd/system/exit-code.service
RUN systemctl enable docker-tests.service exit-code.service

ENTRYPOINT ["/sbin/init"]
