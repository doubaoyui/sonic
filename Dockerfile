## syntax=docker/dockerfile:1.7

ARG GO_VERSION=1.24

FROM golang:${GO_VERSION}-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /src

ARG GOPROXY=https://proxy.golang.org,direct
ARG GOSUMDB=sum.golang.org
ENV GOPROXY=${GOPROXY}
ENV GOSUMDB=${GOSUMDB}

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
  --mount=type=cache,target=/root/.cache/go-build \
  /bin/sh -c 'for i in 1 2 3; do go mod download && exit 0; echo "go mod download failed, retry $i/3" >&2; sleep 2; done; exit 1'

COPY . .

ARG BUILD_COMMIT=dev
ARG BUILD_TIME=unknown
ARG SONIC_VERSION=dev

RUN --mount=type=cache,target=/go/pkg/mod \
  --mount=type=cache,target=/root/.cache/go-build \
  CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} \
  go build -o /out/sonic \
  -ldflags="-s -w -X github.com/go-sonic/sonic/consts.SonicVersion=${SONIC_VERSION} -X github.com/go-sonic/sonic/consts.BuildCommit=${BUILD_COMMIT} -X github.com/go-sonic/sonic/consts.BuildTime=${BUILD_TIME}" \
  -trimpath .

RUN mkdir -p /out/app \
  && cp /out/sonic /out/app/ \
  && cp -r /src/conf /out/app/ \
  && cp -r /src/resources /out/app/ \
  && cp /src/scripts/docker_init.sh /out/app/


FROM alpine:latest AS prod

COPY --from=builder /out/app/ /app/

RUN apk add --no-cache tzdata ca-certificates \
  && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
  && echo "Asia/Shanghai" > /etc/timezone \
  && sed -i 's/\r$//' /app/docker_init.sh \
  && chmod +x /app/docker_init.sh

VOLUME /sonic
EXPOSE 8080

WORKDIR /sonic
CMD ["/bin/sh", "-c", "/app/docker_init.sh && exec /app/sonic -config /sonic/conf/config.yaml"]
