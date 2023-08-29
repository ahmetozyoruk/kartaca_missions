#build stage
FROM golang:alpine AS builder

RUN apk add --no-cache git
WORKDIR /go/src/app
COPY . .

RUN go get -d -v ./...
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /go/bin/app -v .

#final stage
FROM alpine:latest

RUN addgroup -S app && adduser -S app -G app
COPY --from=builder --chown=app /go/bin/app /app

# USER appdocker volume ls

EXPOSE 8080/tcp

ENTRYPOINT ["/app"] 