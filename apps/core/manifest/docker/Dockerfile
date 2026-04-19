FROM golang:1.24-alpine AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/easymvp-core ./main.go

FROM alpine:3.20

WORKDIR /app
RUN addgroup -S easymvp && adduser -S easymvp -G easymvp
COPY --from=builder /out/easymvp-core /app/easymvp-core
COPY manifest /app/manifest
RUN mkdir -p /app/var/data && chown -R easymvp:easymvp /app

USER easymvp
EXPOSE 8000
ENTRYPOINT ["/app/easymvp-core"]
