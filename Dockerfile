FROM golang:1.15-buster AS builder
ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get update -y && apt-get -y install libpcsclite-dev

COPY . /ingress-yubikey

WORKDIR /ingress-yubikey
RUN go mod download
RUN go build . && strip ./ingress-yubikey

FROM debian:buster-slim
ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get update && apt-get -y install libpcsclite1 && rm -rf /var/lib/apt/lists/*
COPY --from=builder /ingress-yubikey/ingress-yubikey /usr/local/bin/ingress-yubikey
ENTRYPOINT ["ingress-yubikey"]
