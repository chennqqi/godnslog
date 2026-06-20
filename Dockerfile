# build frontend
FROM node:24.13.0-alpine as frontend-builder
WORKDIR /app
COPY frontend-next /app
RUN npm config set registry https://registry.npmmirror.com && npm install
RUN npm run build

# build backend
FROM golang:1.25-alpine AS backend-builder

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

# build app
FROM alpine:3.20

RUN apk add --no-cache -U tzdata ca-certificates libcap wget && \
	update-ca-certificates

RUN mkdir -p /app

COPY --from=backend-builder /go/bin/godnslog /app/godnslog
COPY --from=frontend-builder /app/dist /app/dist

ARG UID=1000
ARG GID=1000

RUN addgroup -g $GID -S app && adduser -u $UID -S -g app app && \
  chown -R app:app /app && \
  setcap cap_net_bind_service=eip /app/godnslog

WORKDIR /app
USER app

EXPOSE 8080
EXPOSE 53/UDP 53/TCP

HEALTHCHECK --interval=20s --timeout=3s --start-period=15s --retries=3 \
  CMD wget -qO- http://localhost:8080/api/v2/health || exit 1

ENTRYPOINT [ "/app/godnslog" ]
