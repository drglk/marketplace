FROM golang:1.24.5-alpine3.22 AS build
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod tidy

COPY . .
RUN go build -o ./build/api ./cmd/app

FROM alpine:3
WORKDIR /app
COPY --from=build /app/build/api /app/api
COPY --from=build /app/config ./config
RUN apk add --no-cache curl

ARG CONFIG_PATH=./config/config.yaml
ENV CONFIG_PATH=${CONFIG_PATH}



ENTRYPOINT ["/app/api"]