FROM alpine:latest

ADD bin/server /server
ADD etc/config.yaml /config.yaml

EXPOSE 53/tcp 53/udp

ENTRYPOINT ["/server", "-c", "/config.yaml"]