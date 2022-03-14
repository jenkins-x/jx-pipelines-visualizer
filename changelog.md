### Linux

```shell
curl -L https://github.com/jenkins-x/jx-pipelines-visualizer/releases/download/v1.8.0/jx-pipelines-visualizer-linux-amd64.tar.gz | tar xzv 
sudo mv jx-pipelines-visualizer /usr/local/bin
```

### macOS

```shell
curl -L  https://github.com/jenkins-x/jx-pipelines-visualizer/releases/download/v1.8.0/jx-pipelines-visualizer-darwin-amd64.tar.gz | tar xzv
sudo mv jx-pipelines-visualizer /usr/local/bin
```
## Changes

### New Features

* By default listen on all namespaces (Damian Kęska)
* Support for multiple namespaces with preserving backwards compatibility (Damian Kęska)

### Bug Fixes

* Remove todos (Damian Kęska)

### Code Refactoring

* Code review fixes to fix Tekton endpoint https://github.com/jenkins-x/jx-pipelines-visualizer/pull/145#discussion_r825917477 (Damian Kęska)

### Documentation

* Update docs about multiple namespaces (Damian Kęska)
