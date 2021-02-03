### Linux

```shell
curl -L https://github.com/jenkins-x/jx-pipelines-visualizer/releases/download/v1.1.5/jx-pipelines-visualizer-linux-amd64.tar.gz | tar xzv 
sudo mv jx-pipelines-visualizer /usr/local/bin
```

### macOS

```shell
curl -L  https://github.com/jenkins-x/jx-pipelines-visualizer/releases/download/v1.1.5/jx-pipelines-visualizer-darwin-amd64.tar.gz | tar xzv
sudo mv jx-pipelines-visualizer /usr/local/bin
```
## Changes

### Bug Fixes

* handle live logs for repos with underscore (Vincent Behar) [#94](https://github.com/jenkins-x/jx-pipelines-visualizer/issues/94) 
* support repos with underscore (Vincent Behar) [#91](https://github.com/jenkins-x/jx-pipelines-visualizer/issues/91) 

### Chores

* add file for changelog (James Strachan)

### Issues

* [#94](https://github.com/jenkins-x/jx-pipelines-visualizer/issues/94) In progress logs not available for repos with underscore
* [#91](https://github.com/jenkins-x/jx-pipelines-visualizer/issues/91) Error: Archived logs not found in the long term storage for repos with underscores in name 
