FROM golang:1.26-alpine AS build

WORKDIR /src

COPY go.mod ./
COPY cmd ./cmd
COPY internal ./internal

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/conveyance ./cmd/conveyance

FROM scratch

COPY --from=build /out/conveyance /conveyance

USER 65532:65532

ENTRYPOINT ["/conveyance"]
