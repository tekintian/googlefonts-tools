FROM --platform=$BUILDPLATFORM golang:1.22-alpine AS builder

ARG TARGETOS
ARG TARGETARCH
ARG APPVERSION=dev
ARG BUILD_TAGS

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN if [ -n "$BUILD_TAGS" ]; then \
      CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
      go build -tags "$BUILD_TAGS" -ldflags="-s -w -X main.AppVersion=${APPVERSION}" -o /googlefonts-tools . ; \
    else \
      CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
      go build -ldflags="-s -w -X main.AppVersion=${APPVERSION}" -o /googlefonts-tools . ; \
    fi

FROM alpine:3.20

WORKDIR /app
COPY --from=builder /googlefonts-tools .
COPY config.ini .

RUN mkdir -p storage/db storage/cache storage/fonts storage/zip

EXPOSE 8000

ENTRYPOINT ["./googlefonts-tools"]
CMD ["-s"]