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

FROM alpine:3.20 AS envsubst-builder
RUN apk add --no-cache gcc make musl-dev git && \
    git clone --depth 1 https://github.com/tekintian/envsubst.git /tmp/envsubst-src && \
    cd /tmp/envsubst-src && \
    make && \
    mv envsubst /usr/local/bin/envsubst

FROM alpine:3.20

WORKDIR /app
COPY --from=builder /googlefonts-tools .
COPY --from=envsubst-builder /usr/local/bin/envsubst /usr/local/bin/envsubst
COPY storage/config.ini /app/config.ini.tpl

RUN mkdir -p storage/db storage/cache storage/fonts storage/zip

COPY <<'ENTRYEOF' /app/entrypoint.sh
#!/bin/sh
CONFIG="storage/config.ini"
TEMPLATE="/app/config.ini.tpl"

if [ -f "$CONFIG" ]; then
  envsubst < "$CONFIG" > "$CONFIG.tmp" && mv "$CONFIG.tmp" "$CONFIG"
elif [ -f "$TEMPLATE" ]; then
  envsubst < "$TEMPLATE" > "$CONFIG"
else
  echo "[server]" > "$CONFIG"
  echo "host=${GF_SERVER_HOST:-localhost}" >> "$CONFIG"
  echo "port=${GF_SERVER_PORT:-8000}" >> "$CONFIG"
  echo "" >> "$CONFIG"
  echo "[database]" >> "$CONFIG"
  echo "driver=${GF_DB_DRIVER:-sqlite}" >> "$CONFIG"
  echo "dsn=${GF_DB_DSN:-}" >> "$CONFIG"
  echo "" >> "$CONFIG"
  echo "[notify]" >> "$CONFIG"
  echo "dingtalk_webhook=${GF_DINGTALK_WEBHOOK:-}" >> "$CONFIG"
  echo "wechat_webhook=${GF_WECHAT_WEBHOOK:-}" >> "$CONFIG"
  echo "smtp_host=${GF_SMTP_HOST:-}" >> "$CONFIG"
  echo "smtp_port=${GF_SMTP_PORT:-25}" >> "$CONFIG"
  echo "smtp_from=${GF_SMTP_FROM:-}" >> "$CONFIG"
  echo "smtp_password=${GF_SMTP_PASSWORD:-}" >> "$CONFIG"
  echo "smtp_to=${GF_SMTP_TO:-}" >> "$CONFIG"
fi

exec ./googlefonts-tools "$@"
ENTRYEOF
RUN chmod +x /app/entrypoint.sh

EXPOSE 8000
VOLUME ["/app/storage"]

ENTRYPOINT ["/app/entrypoint.sh"]
CMD ["-s"]