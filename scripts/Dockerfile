## syntax=docker/dockerfile:1.7

ARG GO_VERSION=1.24

FROM golang:${GO_VERSION}-alpine as builder
RUN apk --no-cache add git ca-certificates

COPY . /go/src/github.com/go-sonic/sonic/
WORKDIR /go/src/github.com/go-sonic/sonic

ARG GOPROXY=https://proxy.golang.org,direct
ARG GOSUMDB=sum.golang.org
ENV GOPROXY=${GOPROXY}
ENV GOSUMDB=${GOSUMDB}

ARG BUILD_COMMIT
ARG BUILD_TIME
ARG SONIC_VERSION

    
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    /bin/sh -c 'for i in 1 2 3; do go mod download && exit 0; echo \"go mod download failed, retry $i/3\" >&2; sleep 2; done; exit 1'

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux && \
go build -o sonic -ldflags="-s -w -X github.com/go-sonic/sonic/consts.SonicVersion=${SONIC_VERSION} -X github.com/go-sonic/sonic/consts.BuildCommit=${BUILD_COMMIT} -X github.com/go-sonic/sonic/consts.BuildTime=${BUILD_TIME}" -trimpath .

RUN mkdir -p /app/conf && \
    mkdir /app/resources && \
    cp -r /go/src/github.com/go-sonic/sonic/sonic /app/ && \
    cp -r /go/src/github.com/go-sonic/sonic/conf /app/ && \
    cp -r /go/src/github.com/go-sonic/sonic/resources /app/ && \
    cp /go/src/github.com/go-sonic/sonic/scripts/docker_init.sh /app/




FROM alpine:latest as prod

COPY --from=builder /app/ /app/

RUN  apk add --no-cache tzdata \
    && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo "Asia/Shanghai" > /etc/timezone \
    && sed -i 's/\r$//' /app/docker_init.sh \
    && chmod +x /app/docker_init.sh

VOLUME /sonic
EXPOSE 8080

WORKDIR /sonic
CMD /app/docker_init.sh && /app/sonic -config /sonic/conf/config.yaml
