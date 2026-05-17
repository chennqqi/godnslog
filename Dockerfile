# build frontend
FROM node:24.13.0-alpine as frontend-builder
WORKDIR /app
COPY frontend-next /app
RUN npm config set registry https://registry.npmmirror.com && npm install
RUN npm run build

# build backend
FROM golang:1.22-alpine as backend-builder

RUN echo "https://mirror.tuna.tsinghua.edu.cn/alpine/v3.12/main" > /etc/apk/repositories

COPY models /src/godnslog/models
COPY server /src/godnslog/server
COPY cache /src/godnslog/cache
COPY internal /src/godnslog/internal
COPY cmd /src/godnslog/cmd
COPY cli /src/godnslog/cli
COPY migration /src/godnslog/migration
COPY *.go go.mod go.sum /src/godnslog/
WORKDIR /src/godnslog
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags="-w -s" -o /go/bin/godnslog

# build app
FROM alpine:3.13.5

RUN apk add --no-cache -U tzdata ca-certificates libcap && \
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

ENTRYPOINT [ "/app/godnslog" ]
