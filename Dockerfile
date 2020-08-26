FROM node:lts-alpine as frontend-builder
WORKDIR /app
COPY frontend/* .
RUN npm install
RUN yarn build


FROM golang:alpine as backend-builder

RUN go env -w GOPROXY=https://goproxy.cn
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags=¡±-w -s¡± -o /go/bin/godnslog

COPY . $GOPATH/src/mypackage/myapp/
WORKDIR $GOPATH/src/mypackage/myapp/

FROM alpine:3.8

# tinghua mirror
RUN echo "https://mirror.tuna.tsinghua.edu.cn/alpine/v3.8/main" > /etc/apk/repositories

COPY --from backend-builder /go/bin/godnslog /app/godnslog
COPY --from frontend-builder /app/dist /app/dist

ENTRYPOINT /app/godnslog