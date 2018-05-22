FROM alpine:3.6

RUN adduser -D netperf-operator
USER netperf-operator

ADD tmp/_output/bin/netperf-operator /usr/local/bin/netperf-operator
