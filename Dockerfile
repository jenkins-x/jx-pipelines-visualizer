FROM alpine:3.12

COPY ./webr /app/
COPY ./build/linux/jx-pipelines-visualizer /app/

ENTRYPOINT ["/app/jx-pipelines-visualizer"]