ARG CADDY_VERSION=2.6.0
ARG GO_VERSION=1.22.1
ARG BUILD_DATE
ARG VCS_REF

FROM golang:${GO_VERSION} AS builder
ARG SUPERVISORD_VERSION=v0.7.3
ARG BUILD_DATE
ARG VCS_REF

#
# Build our go apps ( tokenvendor + supervisord)
#
RUN mkdir /app

RUN git clone --depth 1 https://github.com/ochinchina/supervisord.git -b ${SUPERVISORD_VERSION} /tmp/supervisord \
    && cd /tmp/supervisord \
    && GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o /app/supervisord

COPY . /app/
RUN cd /app/auth-token-vendor && GITHUB_SHA=${VCS_REF} GOOS=linux GOARCH=arm64 CGO_ENABLED=0 make tokenvendor/build


FROM --platform=linux/arm64 caddy:${CADDY_VERSION}
ARG BUILD_DATE
ARG VCS_REF

LABEL architecture="linux/arm64" \
    caddy-version=${CADDY_VERSION} \
    build=${BUILD_DATE} \
    org.opencontainers.image.revision=${VCS_REF} \
    org.opencontainers.image.source="https://github.com/NixM0nk3y/databricks-app-proxy" \
    org.opencontainers.image.created=${BUILD_DATE}

#
RUN echo "==> Installing extra packages..." \
    && apk add --no-cache bash curl jq

# drop in our extra config and scripts
COPY config/Caddyfile /etc/caddy/Caddyfile
COPY config/supervisord.conf /etc/supervisord.conf

# copy in our go apps
COPY --from=builder /app/auth-token-vendor/tokenvendor /app/tokenvendor
COPY --from=builder /app/supervisord /app/supervisord

RUN apk add --no-cache curl && \
    rm -f /var/cache/apk/* /tmp/*

COPY ./config/Caddyfile /etc/caddy/Caddyfile

# start out supervisord process
CMD [ "/app/supervisord","-c","/etc/supervisord.conf" ]
