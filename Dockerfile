FROM golang:1.25-alpine AS builder
WORKDIR /src
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 \
    go build -trimpath -ldflags="-s -w" \
    -o /out/pocket-aide ./cmd/server

FROM gcr.io/distroless/static:nonroot
COPY --from=builder /out/pocket-aide /usr/local/bin/pocket-aide
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/pocket-aide"]
