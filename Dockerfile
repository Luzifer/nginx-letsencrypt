FROM golang:alpine

LABEL maintainer "Knut Ahlers <knut@ahlers.me>"

ENV BUFFER=720h \
    NGINX_CONFIG=/data/config/nginx.conf \
    EMAIL=mail@example.com \
    STORAGE_DIR=/data/ssl

ADD . /go/src/github.com/Luzifer/nginx-letsencrypt
WORKDIR /go/src/github.com/Luzifer/nginx-letsencrypt

RUN set -ex \
 && apk add --update git ca-certificates \
 && go install -ldflags "-X main.version=$(git describe --tags || git rev-parse --short HEAD || echo dev)" \
 && apk del --purge git

EXPOSE 80/tcp 443/tcp

VOLUME ["/data/ssl", "/data/config"]

ENTRYPOINT ["/go/bin/nginx-letsencrypt"]
CMD ["--"]
