FROM golang:1.25-bookworm AS builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /mantis-wallet ./cmd/wallet

FROM gcr.io/distroless/static-debian12
COPY --from=builder /mantis-wallet /mantis-wallet
EXPOSE 50054
CMD ["/mantis-wallet"]
