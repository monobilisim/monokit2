FROM golang:1.25-trixie

WORKDIR /app

COPY . .

RUN apt update
RUN apt install -y build-essential

ENV CGO_ENABLED=1

RUN mkdir -p /etc/mono
RUN cp config/* /etc/mono
# Only compile monokit2 for database creation
RUN make build

CMD [ "make", "test" ]
