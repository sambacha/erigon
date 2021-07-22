# syntax=docker/dockerfile:1
FROM docker.io/library/golang:1.16-alpine3.13 AS builder

RUN apk --no-cache add make gcc g++ linux-headers git bash ca-certificates libgcc libstdc++

WORKDIR /app
ADD . .
RUN make erigon rpcdaemon integration sentry


FROM docker.io/library/alpine:3.13
RUN mkdir -p /var/lib/erigon
VOLUME /var/lib/erigon

RUN apk add --no-cache ca-certificates libgcc libstdc++ tzdata
COPY --from=builder /app/build/bin/* /usr/local/bin/

WORKDIR /var/lib/erigon

RUN adduser -H -u 1000 -g 1000 -D erigon
RUN chown -R erigon:erigon /var/lib/erigon
USER erigon

EXPOSE 8545 8546 30303 30303/udp 30304 30304/udp 8080 9090 6060


ARG BUILD_DATE
ARG VCS_REF
LABEL org.label-schema.build-date=$BUILD_DATE \
      org.label-schema.name="Erigon" \
      org.label-schema.description="Erigon Ethereum Client" \
      org.label-schema.url="https://torquem.ch" \
      org.label-schema.vcs-ref=$VCS_REF \
      org.label-schema.vcs-url="https://github.com/ledgerwatch/erigon.git" \
      org.label-schema.vendor="Torquem" \
      org.label-schema.version=$VERSION \
      org.label-schema.schema-version="1.0"
