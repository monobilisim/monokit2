FROM golang:1.25-trixie

WORKDIR /app

COPY . .

RUN apt update
RUN apt install -y build-essential

ENV CGO_ENABLED=1

RUN mkdir -p /etc/mono
RUN cp config/* /etc/mono
RUN chmod +x scripts/collect-test-artifacts.sh

CMD [ "sh", "-c", "make test-must-run-on-docker; EXIT_CODE=$?; ./scripts/collect-test-artifacts.sh; exit $EXIT_CODE" ]
