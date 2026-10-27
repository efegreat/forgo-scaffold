# syntax=docker/dockerfile:1

# One image holds every deployable, and each container runs one of them: /usr/local/bin/http, grpc, worker,
# cron or cli. Run the binary itself as the entrypoint, in exec form, so that it receives SIGTERM and drains.
#
# Until forgo and forgo-adapter are released, go.mod reads them from checkouts next to this one, which the
# build takes as the contexts named forgo and forgo-adapter (compose.yaml passes them):
#
#   docker build --build-context forgo=../forgo --build-context forgo-adapter=../forgo-adapter -t example .
#
# Once they are released, drop the two COPY --from lines below and the replace block of go.mod.
FROM golang:1.27 AS build
WORKDIR /src/forgo-scaffold
COPY --from=forgo . /src/forgo
COPY --from=forgo-adapter . /src/forgo-adapter
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/ ./cmd/...

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/ /usr/local/bin/
