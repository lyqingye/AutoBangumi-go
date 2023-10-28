# Build the application from source
FROM golang:1.21.3-alpine3.17 AS build-stage

WORKDIR /app

RUN  apk add git --no-cache

COPY go.mod go.sum ./
RUN go env -w GO111MODULE=on  &&  go mod download

COPY . .

RUN go mod tidy

RUN go build -o ab


# Deploy the application binary into a lean image
FROM alpine:3.17

WORKDIR /

COPY --from=build-stage /app/ab /ab

ENTRYPOINT ["/ab", "bot", "run-ab-bot"]
