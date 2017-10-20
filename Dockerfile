FROM alpine:3.5

ADD ./endpoint-operator /endpoint-operator

ENTRYPOINT ["/endpoint-operator"]