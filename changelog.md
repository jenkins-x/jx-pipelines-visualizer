### Linux

```shell
curl -L https://github.com/jenkins-x/jx-pipelines-visualizer/releases/download/v1.1.6/jx-pipelines-visualizer-linux-amd64.tar.gz | tar xzv 
sudo mv jx-pipelines-visualizer /usr/local/bin
```

### macOS

```shell
curl -L  https://github.com/jenkins-x/jx-pipelines-visualizer/releases/download/v1.1.6/jx-pipelines-visualizer-darwin-amd64.tar.gz | tar xzv
sudo mv jx-pipelines-visualizer /usr/local/bin
```
## Changes

### Bug Fixes

* handle live logs for repos with underscore (Vincent Behar) [#94](https://github.com/jenkins-x/jx-pipelines-visualizer/issues/94) 

### Chores

* upgrade deps (James Strachan)

### Issues

* [#94](https://github.com/jenkins-x/jx-pipelines-visualizer/issues/94) In progress logs not available for repos with underscore
