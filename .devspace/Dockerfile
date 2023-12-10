FROM golang:alpine as go

FROM dustinliu/nvim:latest
COPY --from=go /usr/local/go /usr/local/go
RUN apk update && apk add --no-cache make
