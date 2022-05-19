### Linux

```shell
curl -L https://github.com/jenkins-x/jx-pipelines-visualizer/releases/download/v1.8.2/jx-pipelines-visualizer-linux-amd64.tar.gz | tar xzv 
sudo mv jx-pipelines-visualizer /usr/local/bin
```

### macOS

```shell
curl -L  https://github.com/jenkins-x/jx-pipelines-visualizer/releases/download/v1.8.2/jx-pipelines-visualizer-darwin-amd64.tar.gz | tar xzv
sudo mv jx-pipelines-visualizer /usr/local/bin
```
## Changes

### Bug Fixes

* downgrade ie-proxy to fix darwin build (ankitm123)
