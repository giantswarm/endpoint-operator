FROM alpine:3.8

ADD ./endpoint-operator /endpoint-operator

ENTRYPOINT ["/endpoint-operator"]
