FROM golang:1.23 AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal
COPY migrations ./migrations
COPY pkg ./pkg

RUN CGO_ENABLED=0 go build -o /ratelimiter ./cmd/ratelimiter/main.go

FROM alpine:3.20

COPY --from=build /ratelimiter /ratelimiter

ENTRYPOINT ["/ratelimiter"]