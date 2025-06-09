FROM node:18 as builder

WORKDIR /build

ENV YARN_NETWORK_TIMEOUT=300000
ENV YARN_NETWORK_CONCURRENCY=1
ENV YARN_RETRY_TIMES=3

COPY web/package.json web/yarn.lock ./

RUN yarn config set registry https://registry.yarnpkg.com/ && \
    yarn --frozen-lockfile --network-timeout 300000 || \
    (sleep 3 && yarn --frozen-lockfile --network-timeout 300000)

COPY ./web .
COPY ./VERSION .
RUN DISABLE_ESLINT_PLUGIN='true' VITE_APP_VERSION=$(cat VERSION) npm run build

FROM golang:1.24.2 AS builder2

ENV GO111MODULE=on \
    CGO_ENABLED=1 \
    GOOS=linux \
    GOPROXY=https://proxy.golang.org,direct

WORKDIR /build
ADD go.mod go.sum ./
RUN go mod download || (sleep 3 && go mod download)
COPY . .
COPY --from=builder /build/build ./web/build
RUN go build -ldflags "-s -w -X 'done-hub/common.Version=$(cat VERSION)' -extldflags '-static'" -o done-hub

FROM alpine:latest

RUN apk update && \
    apk upgrade && \
    apk add --no-cache ca-certificates tzdata && \
    update-ca-certificates 2>/dev/null || true

COPY --from=builder2 /build/done-hub /
EXPOSE 3000
WORKDIR /data
ENTRYPOINT ["/done-hub"]
