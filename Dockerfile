# syntax=docker/dockerfile:1

# One image holds every deployable, and each container runs one of them: /usr/local/bin/http, grpc, worker,
# cron or cli. Run the binary itself as the entrypoint, in exec form, so that it receives SIGTERM and drains.
#
#   docker build -t example .
FROM golang:1.27 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/ ./cmd/...

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/ /usr/local/bin/
