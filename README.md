# Jenkins X Pipelines Visualizer

This is a Web UI for Jenkins X, with a clear goal: **visualize the pipelines - and their logs**.

It's a web server written in Go, producing HTML content.

## Features

- very simple
- very fast: get your logs in milliseconds, not seconds. Yes, I'm looking at you, JXUI.
- either retrieve the build logs from the persistent storage (tested with GCS), or stream them from the running pods if the pipeline is still running.
- retrieve the build logs even for the garbage-collected pipelines (the JXUI just returns 404)
- read only. Only requires READ permissions on the JX and Tekton Pipelines CRDs
- URLs backward-compatible with JXUI - so that you can easily swap the JXUI URL for the jx-pipelines-visualizer one in the Lighthouse config, and have Lighthouse set links to jx-pipelines-visualizer in GitHub Pull Requests.

### Out of scope

- Auth - use a reverse-proxy in front or anything else to handle it
- Create/Update/Delete operations. It is meant to be a read-only web UI
- Anything in JX which is not related to the pipelines

## How It Works

It uses the "informer" Kubernetes pattern to keep a local cache of the Jenkins X PipelineActivities, and index them in an in-memory [Bleve](http://blevesearch.com/) index.

It uses part of jx code to retrieve the build logs - mainly the part to stream the build logs from the running pods. It is the same code used by the `jx get build logs` command.
