# build frontend
FROM node:24.13.0-alpine AS frontend-builder
WORKDIR /app
COPY frontend-next/package.json frontend-next/package-lock.json* ./
RUN npm config set registry https://registry.npmmirror.com && npm install
COPY frontend-next ./
RUN npm run build && rm -rf dist/dev/cache

# build backend
FROM golang:1.25-alpine AS backend-builder

ENV GOPROXY=https://goproxy.cn,direct

RUN apk add --no-cache build-base git musl-dev

COPY models /src/godnslog/models
COPY server /src/godnslog/server
COPY cache /src/godnslog/cache
COPY internal /src/godnslog/internal
COPY cmd /src/godnslog/cmd
COPY cli /src/godnslog/cli
COPY migration /src/godnslog/migration
COPY templates /src/godnslog/templates
COPY *.go go.mod go.sum /src/godnslog/
WORKDIR /src/godnslog
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags="-w -s" -o /go/bin/godnslog

# final image: Alpine base with Go binary + frontend static files
# NOTE: node:24.13.0-alpine is used instead of plain alpine:3.21
# because the latter may not be available behind certain proxies.
FROM node:24.13.0-alpine

RUN apk add --no-cache -U tzdata ca-certificates libcap wget tini && \
	update-ca-certificates

RUN mkdir -p /app/frontend/dist /app

COPY --from=backend-builder /go/bin/godnslog /app/godnslog
COPY --from=frontend-builder /app/dist/standalone /app/frontend
COPY --from=frontend-builder /app/dist/static /app/frontend/dist/static

ARG UID=1001
ARG GID=1001

RUN addgroup -g $GID -S app && adduser -u $UID -S -g app app && \
  chown -R app:app /app && \
  setcap cap_net_bind_service=eip /app/godnslog

WORKDIR /app
USER app

EXPOSE 8080
EXPOSE 80
EXPOSE 443
EXPOSE 53/UDP 53/TCP

HEALTHCHECK --interval=20s --timeout=3s --start-period=15s --retries=3 \
  CMD wget -qO- http://localhost:8080/api/v2/health || exit 1

COPY deploy/docker/entrypoint.sh /app/entrypoint.sh

ENTRYPOINT ["/sbin/tini", "--"]
CMD ["/app/entrypoint.sh"]
