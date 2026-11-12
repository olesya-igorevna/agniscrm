FROM golang:1.26-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY cmd/ ./cmd/
COPY internal/ ./internal/
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/agniscrm ./cmd/server

FROM scratch

WORKDIR /app
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /out/agniscrm /app/agniscrm

ENV HTTP_ADDR=:8080 \
    GIN_MODE=release

USER 65532:65532
EXPOSE 8080
ENTRYPOINT ["/app/agniscrm"]
