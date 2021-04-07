### Linux

```shell
curl -L https://github.com/jenkins-x/jx-pipelines-visualizer/releases/download/v1.3.1/jx-pipelines-visualizer-linux-amd64.tar.gz | tar xzv 
sudo mv jx-pipelines-visualizer /usr/local/bin
```

### macOS

```shell
curl -L  https://github.com/jenkins-x/jx-pipelines-visualizer/releases/download/v1.3.1/jx-pipelines-visualizer-darwin-amd64.tar.gz | tar xzv
sudo mv jx-pipelines-visualizer /usr/local/bin
```
## Changes

### New Features

* open pipeline trace URL (Vincent Behar)

### Bug Fixes

* don't fail if the trace button is not there (Vincent Behar) [#107](https://github.com/jenkins-x/jx-pipelines-visualizer/issues/107) 
* release pipeline (James Strachan)
* only display trace link if there is a trace (Vincent Behar)

### Issues

* [#107](https://github.com/jenkins-x/jx-pipelines-visualizer/issues/107) jx3: not seeing logs after clicking on build in UI
