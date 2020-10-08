FROM alpine:3.12

RUN apk add --no-cache ca-certificates \
 && adduser -D -u 1000 jx

COPY ./web/static /app/web/static
COPY ./web/templates /app/web/templates
COPY ./build/linux/jx-pipelines-visualizer /app/

WORKDIR /app
USER 1000

ENTRYPOINT ["/app/jx-pipelines-visualizer"]