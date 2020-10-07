FROM alpine:3.12

COPY ./web /app/
COPY ./build/linux/jx-pipelines-visualizer /app/

WORKDIR /app

ENTRYPOINT ["/app/jx-pipelines-visualizer"]