FROM scratch

ADD bin/server /dist/bin/server
ADD etc /dist/etc
ADD ca-certificates.crt /etc/ssl/certs/

EXPOSE 53/tcp 53/udp

ENTRYPOINT ["/dist/bin/server", "-c", "/dist/etc/config.yaml"]