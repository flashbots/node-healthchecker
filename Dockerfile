# stage: build ---------------------------------------------------------

FROM golang:1.22-alpine as build

RUN apk add --no-cache gcc musl-dev linux-headers

WORKDIR /go/src/github.com/flashbots/node-healthchecker

COPY go.* ./
RUN go mod download

COPY . .

RUN go build -o bin/node-healthchecker -ldflags "-s -w" github.com/flashbots/node-healthchecker/cmd

# stage: run -----------------------------------------------------------

FROM alpine

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=build /go/src/github.com/flashbots/node-healthchecker/bin/node-healthchecker ./node-healthchecker

ENTRYPOINT ["/app/node-healthchecker"]
