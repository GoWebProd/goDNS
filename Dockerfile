FROM scratch

ADD bin/server /dist/bin/server
ADD etc /dist/etc
RUN chmod +x /dist/bin/server
ADD ca-certificates.crt /etc/ssl/certs/

EXPOSE 53/tcp 53/udp
VOLUME /dist/etc/config.yaml

WORKDIR /dist/
ENTRYPOINT "./bin/server"