FROM golang:1.13-alpine AS src

RUN apk update && apk upgrade; \
    apk add build-base

WORKDIR /app
COPY go.mod go.sum *.go ./
RUN pwd

RUN GOOS=linux go build -o ./app;

# Final image, no source code
FROM alpine:latest

RUN apk update && apk upgrade; \
    apk add build-base

WORKDIR .
COPY --from=src /app .

EXPOSE 8080

# Run Go Binary
CMD ./app