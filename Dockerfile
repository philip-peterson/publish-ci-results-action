FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /publish-test-results .

FROM alpine:latest
RUN apk add --no-cache ca-certificates
COPY --from=builder /publish-test-results /publish-test-results
ENTRYPOINT ["/publish-test-results"]
