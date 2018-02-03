FROM alpine:latest

ADD bin/server /server
ADD etc/config.yaml /config.yaml
ADD ca-certificates.crt /etc/ssl/certs/
RUN mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2

EXPOSE 53/tcp 53/udp

CMD ["/server", "-c", "/config.yaml"]