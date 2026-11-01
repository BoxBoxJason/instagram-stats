FROM docker.io/golang:1.27.2-alpine@sha256:f92b6ef800e499660581efdabdf25d9d817a9d124eaf900924f0504e7e27e12d AS build

WORKDIR /app

COPY . .

ENV GO111MODULE=on \
    CGO_ENABLED=0

RUN apk add --no-cache make git && \
  make build

FROM alpine:3.24.2@sha256:294b683cb724975bec92580e1e685676bd4b50bda910ddb8c51d4cabeaec77e6 AS security_provider

RUN apk add --no-cache ca-certificates \
    && addgroup -S instagram-stats \
    && adduser -S instagram-stats -G instagram-stats \
    && mkdir /work \
    && chown instagram-stats:instagram-stats /work

FROM scratch

COPY --from=security_provider /etc/passwd /etc/passwd
COPY --from=security_provider /etc/group /etc/group
# The default result folder, Results, is relative to the current folder: it must be writable.
COPY --from=security_provider --chown=instagram-stats:instagram-stats /work /work

WORKDIR /work
USER instagram-stats

COPY --from=build /app/bin/instagram-stats /usr/local/bin/instagram-stats

ENTRYPOINT [ "/usr/local/bin/instagram-stats" ]
