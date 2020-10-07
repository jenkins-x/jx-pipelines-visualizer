FROM alpine:3.12

COPY ./build/linux/jx-pipelines-visualizer /bin/jx-pipelines-visualizer
ENTRYPOINT ["/bin/jx-pipelines-visualizer"]