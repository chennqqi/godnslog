# build frontend
FROM node:24.13.0-alpine AS frontend-builder
WORKDIR /app
COPY frontend-next/package.json frontend-next/package-lock.json* ./
RUN npm config set registry https://registry.npmmirror.com && npm install
COPY frontend-next ./
RUN npm run build

# build backend
FROM golang:1.25-alpine AS backend-builder

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

# final image: Node runtime for Next.js + Go binary
FROM node:24.13.0-alpine

RUN apk add --no-cache -U tzdata ca-certificates libcap wget tini && \
	update-ca-certificates

RUN mkdir -p /app/frontend /app

COPY --from=backend-builder /go/bin/godnslog /app/godnslog
COPY --from=frontend-builder /app/dist /app/frontend/dist
COPY --from=frontend-builder /app/package.json /app/frontend/package.json
COPY --from=frontend-builder /app/next.config.js /app/frontend/next.config.js
COPY --from=frontend-builder /app/node_modules /app/frontend/node_modules

ARG UID=1001
ARG GID=1001

RUN addgroup -g $GID -S app && adduser -u $UID -S -g app app && \
  chown -R app:app /app && \
  setcap cap_net_bind_service=eip /app/godnslog

WORKDIR /app
USER app

ENV GODNSLOG_API_URL=http://localhost:8080

EXPOSE 8080
EXPOSE 3000
EXPOSE 53/UDP 53/TCP

HEALTHCHECK --interval=20s --timeout=3s --start-period=15s --retries=3 \
  CMD wget -qO- http://localhost:8080/api/v2/health || exit 1

# Copy entrypoint script (ensure executable bit is set in source)
COPY deploy/docker/entrypoint.sh /app/entrypoint.sh

# Start Go backend and Next.js frontend with tini for proper signal handling
ENTRYPOINT ["/sbin/tini", "--"]
CMD ["/app/entrypoint.sh"]
