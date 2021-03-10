FROM golang:1.16.2 AS builder

COPY . .
RUN GOPATH= CGO_ENABLED=0 go build -o /bin/kayentactl

FROM alpine:3.12

RUN apk add --no-cache ca-certificates
COPY --from=builder /bin/kayentactl /kayentactl
ENTRYPOINT ["./kayentactl"]
