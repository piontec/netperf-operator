FROM golang:1.10.1 as builder
RUN curl https://github.com/golang/dep/releases/download/v0.4.1/dep-linux-amd64 -L -o /usr/bin/dep
RUN chmod +x /usr/bin/dep
ENV GOPATH=/go
WORKDIR /go/src/github.com/piontec/netperf-operator
COPY Gopkg.* ./
COPY pkg ./pkg
COPY cmd ./cmd
RUN dep ensure
RUN pwd
RUN ls -la
RUN CGO_ENABLED=0 GOOS=linux go build -o netperf-operator cmd/netperf-operator/main.go

FROM alpine:3.6  
WORKDIR /
COPY --from=builder /go/src/github.com/piontec/netperf-operator/netperf-operator /
CMD ["/netperf-operator"]  