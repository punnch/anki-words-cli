FROM golang:1.25-bookworm AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /app/cmd/ankiwords/exe /app/cmd/ankiwords

FROM alpine:3.23.4 AS runner
WORKDIR /app
RUN apk add --no-cache ca-certificates
COPY --from=builder /app/cmd/ankiwords/exe /app/exe
CMD ["/app/exe"]
