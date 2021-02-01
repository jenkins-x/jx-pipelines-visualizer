### Linux

```shell
curl -L https://github.com/jenkins-x/jx-pipelines-visualizer/releases/download/v1.1.4/jx-pipelines-visualizer-linux-amd64.tar.gz | tar xzv 
sudo mv jx-pipelines-visualizer /usr/local/bin
```

### macOS

```shell
curl -L  https://github.com/jenkins-x/jx-pipelines-visualizer/releases/download/v1.1.4/jx-pipelines-visualizer-darwin-amd64.tar.gz | tar xzv
sudo mv jx-pipelines-visualizer /usr/local/bin
```
## Changes

### Bug Fixes

* support repos with underscore (Vincent Behar) [#91](https://github.com/jenkins-x/jx-pipelines-visualizer/issues/91) 

### Chores

* add file for changelog (James Strachan)
* upgrade pipelines (James Strachan)

### Issues

* [#91](https://github.com/jenkins-x/jx-pipelines-visualizer/issues/91) Error: Archived logs not found in the long term storage for repos with underscores in name 
