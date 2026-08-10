FROM golang:1.26-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/conveyance ./cmd/conveyance
RUN mkdir -p /runtime/app \
    && cp /out/conveyance /runtime/app/conveyance

FROM scratch

COPY --from=build --chown=65532:65532 /runtime/app /app

WORKDIR /app

USER 65532:65532

ENTRYPOINT ["/app/conveyance"]
