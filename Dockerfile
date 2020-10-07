FROM alpine:3.12

COPY ./web /app/web
COPY ./build/linux/jx-pipelines-visualizer /app/

WORKDIR /app

ENTRYPOINT ["/app/jx-pipelines-visualizer"]