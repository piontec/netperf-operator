FROM alpine:3.6

RUN adduser -D netperf-operator
USER netperf-operator

ADD netperf-operator /usr/local/bin/netperf-operator
CMD ["/usr/local/bin/netperf-operator"]
