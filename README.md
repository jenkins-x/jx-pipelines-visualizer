# Jenkins X Pipelines Visualizer

This is a Web UI for [Jenkins X](https://jenkins-x.io/), with a clear goal: **visualize the pipelines - and their logs**.

See the [announcement blog post](https://jenkins-x.io/blog/2020/09/23/jx-pipelines-visualizer/) for more details and a demo.


## Current Status

This project has been started in September 2020 and shared after a couple of hours of work. It's working - and is deployed and used at [Dailymotion](https://www.dailymotion.com/) - even if the UI is very basic for the moment.

Note that it is being used on GKE with logs stored in a GCS bucket.

### Roadmap

- Improve the UI: visualize the pipeline (stages/steps), group logs per stage/step, and so on
- Split the API and the UI, to make it easier to iterate on the UI

## Features

- very simple
- very fast: get your logs in milliseconds, not seconds. Yes, I'm looking at you, JXUI.
- either retrieve the build logs from the persistent storage (tested with GCS), or stream them from the running pods if the pipeline is still running.
- retrieve the build logs even for the garbage-collected pipelines (the JXUI just returns 404)
- read-only. Only requires READ permissions on the JX and Tekton Pipelines CRDs
- URLs backward-compatible with JXUI - so that you can easily swap the JXUI URL for the jx-pipelines-visualizer one in the Lighthouse config, and have Lighthouse set links to jx-pipelines-visualizer in GitHub Pull Requests.

### Out of scope

- Auth - use a reverse-proxy in front or anything else to handle it
  - for example [Vouch and Okta](https://medium.com/@vbehar/how-to-protect-a-kubernetes-ingress-behind-okta-with-nginx-91e279e06009)
  - or [dex](https://github.com/dexidp/dex), [oauth2-proxy](https://github.com/oauth2-proxy/oauth2-proxy), ...
  - or [nginx basic-auth](https://kubernetes.github.io/ingress-nginx/user-guide/nginx-configuration/annotations/#authentication) - if you are using the nginx ingress controller
- Create/Update/Delete operations. It is meant to be a read-only web UI
- Anything in JX which is not related to the pipelines

## Installation

### With Jenkins X v3

It's already installed by default with Jenkins X v3. If its not (e.g. you have an older git repository) - add the following to your `helmfile.yaml` file in the `releases:` section:

```yaml
releases:
- chart: jx3/jx-pipelines-visualizer
```

By default an ingress is created to access the UI using basic authentication. See the [documentation for how to access it](https://jenkins-x.io/docs/v3/guides/ui/#pipeline-visualizer)

You can see the default values here: <https://github.com/jenkins-x/jxr-versions/tree/master/charts/jx3/jx-pipelines-visualizer>

### With Jenkins X v2

In the Git repository for your dev environment:

- Update the `env/requirements.yaml` file with the following:
  ```
  - name: jx-pipelines-visualizer
    repository: https://storage.googleapis.com/jenkinsxio/charts/
    version: 0.0.30
  ```
- Create a new file `env/jx-pipelines-visualizer/values.tmpl.yaml` with the following content:
  ```
  {{- if .Requirements.storage.logs.enabled }}
  config:
    archivedLogsURLTemplate: >-
      {{ .Requirements.storage.logs.url }}{{`/jenkins-x/logs/{{.Owner}}/{{.Repository}}/{{if hasPrefix .Branch "pr"}}{{.Branch | upper}}{{else}}{{.Branch}}{{end}}/{{.Build}}.log`}}
  {{- end }}

  gitSecretName: ""

  ingress:
    enabled: true
    hosts:
      - pipelines{{.Requirements.ingress.namespaceSubDomain}}{{.Requirements.ingress.domain}}
    {{- if .Requirements.ingress.tls.enabled }}
    tls:
      enabled: true
      secrets:
        # re-use the existing tls secret managed by jx
        {{- if .Requirements.ingress.tls.production }}
        tls-{{ .Requirements.ingress.domain | replace "." "-" }}-p: {}
        {{- else }}
        tls-{{ .Requirements.ingress.domain | replace "." "-" }}-s: {}
        {{- end }}
    {{- end }}
    annotations:
      kubernetes.io/ingress.class: nginx
  ```
  
  This will expose the UI at `pipelines.your.domain.tld` - without any auth. You can add [basic auth by appending a few additional annotations](https://kubernetes.github.io/ingress-nginx/user-guide/nginx-configuration/annotations/#authentication) - re-using the Jenkins X Auth Secret:

  ```
  nginx.ingress.kubernetes.io/auth-type: basic
  nginx.ingress.kubernetes.io/auth-secret: jx-basic-auth
  ```
  
- If you want [Lighthouse](https://github.com/jenkins-x/lighthouse) to add links to your jx-pipelines-visualizer instance from your Pull/Merge Request checks, update the `env/lighthouse-jx/values.tmpl.yaml` file and add the following:
  ```
  env:
    LIGHTHOUSE_REPORT_URL_BASE: "https://pipelines{{.Requirements.ingress.namespaceSubDomain}}{{.Requirements.ingress.domain}}"
  ```

### With Helm v3

```
$ helm repo add jx https://storage.googleapis.com/jenkinsxio/charts/
$ helm install jx-pipelines-visualizer jx/jx-pipelines-Visualizer
```

### With Helm v2

```
$ helm repo add jx https://storage.googleapis.com/jenkinsxio/charts/
$ helm repo update
$ helm install --name jx-pipelines-visualizer jx/jx-pipelines-visualizer
```

## Configuration

See the [values.yaml](charts/jx-pipelines-visualizer/values.yaml) file for the configuration.

If you are not using the Helm Chart, the binary is using CLI flags only - no config files. You can run `jx-pipelines-visualizer -h` to see all the flags.

## Running locally

```
go run cmd/server/main.go
```

## How It Works

It uses the "informer" Kubernetes pattern to keep a local cache of the Jenkins X PipelineActivities, and index them in an in-memory [Bleve](http://blevesearch.com/) index.

It uses part of jx code to retrieve the build logs - mainly the part to stream the build logs from the running pods. It is the same code used by the `jx get build logs` command.

## Credits

Thanks to [Dailymotion](https://www.dailymotion.com/) for creating the [original repository](https://github.com/dailymotion/jx-pipelines-visualizer) and then donate it to the Jenkins X project.

