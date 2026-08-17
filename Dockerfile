FROM --platform=$BUILDPLATFORM golang:1.22-alpine AS builder

ARG TARGETOS
ARG TARGETARCH
ARG APPVERSION=dev

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -ldflags="-s -w -X main.AppVersion=${APPVERSION}" -o /googlefonts-tools .

FROM ghcr.io/tekintian/alpine:3.20

WORKDIR /app
COPY --from=builder /googlefonts-tools .
COPY config.ini .

RUN mkdir -p storage/db storage/cache storage/fonts storage/zip

EXPOSE 8000

ENTRYPOINT ["./googlefonts-tools"]
CMD ["-s"]