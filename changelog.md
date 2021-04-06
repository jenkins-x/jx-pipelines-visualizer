### Linux

```shell
curl -L https://github.com/jenkins-x/jx-pipelines-visualizer/releases/download/v1.3.0/jx-pipelines-visualizer-linux-amd64.tar.gz | tar xzv 
sudo mv jx-pipelines-visualizer /usr/local/bin
```

### macOS

```shell
curl -L  https://github.com/jenkins-x/jx-pipelines-visualizer/releases/download/v1.3.0/jx-pipelines-visualizer-darwin-amd64.tar.gz | tar xzv
sudo mv jx-pipelines-visualizer /usr/local/bin
```
## Changes

### New Features

* open pipeline trace URL (Vincent Behar)
* support archived pipelineruns (Vincent Behar)

### Bug Fixes

* release pipeline (James Strachan)
* only display trace link if there is a trace (Vincent Behar)
