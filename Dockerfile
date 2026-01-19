FROM golang:1.25-trixie

WORKDIR /app

RUN apt-get update && \
    apt-get install -y --no-install-recommends \
        build-essential && \
    rm -rf /var/lib/apt/lists/*

ENV CGO_ENABLED=1

COPY go.mod go.sum ./

RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .

RUN mkdir -p /etc/mono && \
    cp config/* /etc/mono && \
    chmod +x scripts/collect-test-artifacts.sh

CMD [ "sh", "-c", "make test-must-run-on-docker; EXIT_CODE=$?; ./scripts/collect-test-artifacts.sh; exit $EXIT_CODE" ]
