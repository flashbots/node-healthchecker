# stage: build ---------------------------------------------------------

FROM golang:1.22-alpine as build

RUN apk add --no-cache gcc musl-dev linux-headers

WORKDIR /go/src/github.com/flashbots/node-healthchecker

COPY go.* ./
RUN go mod download

COPY . .

RUN SOURCE_DATE_EPOCH=0 CGO_ENABLED=0 go build \
    -trimpath \
    -ldflags "-s -w -buildid=" \
    -o bin/node-healthchecker \
    github.com/flashbots/node-healthchecker/cmd

# stage: run -----------------------------------------------------------

FROM alpine

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=build /go/src/github.com/flashbots/node-healthchecker/bin/node-healthchecker ./node-healthchecker

ENTRYPOINT ["/app/node-healthchecker"]
