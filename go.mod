module github.com/jenkins-x/jx-pipelines-visualizer

require (
	cloud.google.com/go v0.76.0 // indirect
	github.com/Masterminds/sprig/v3 v3.2.0
	github.com/RoaringBitmap/roaring v0.5.5 // indirect
	github.com/blevesearch/bleve v1.0.14
	github.com/glycerine/go-unsnap-stream v0.0.0-20210130063903-47dfef350d96 // indirect
	github.com/golang/snappy v0.0.2 // indirect
	github.com/gorilla/mux v1.8.0
	github.com/jenkins-x/jx-api/v4 v4.0.24
	github.com/jenkins-x/jx-helpers/v3 v3.0.74
	github.com/jenkins-x/jx-pipeline v0.0.102
	github.com/mitchellh/go-homedir v1.1.0
	github.com/rickb777/date v1.13.0
	github.com/rs/xid v1.2.1
	github.com/sirupsen/logrus v1.7.0
	github.com/subchord/go-sse v1.0.1
	github.com/tektoncd/pipeline v0.20.0
	github.com/tinylib/msgp v1.1.5 // indirect
	github.com/unrolled/render v1.0.3
	github.com/urfave/negroni/v2 v2.0.2
	github.com/willf/bitset v1.1.11 // indirect
	golang.org/x/crypto v0.0.0-20201221181555-eec23a3978ad // indirect
	golang.org/x/oauth2 v0.0.0-20210201163806-010130855d6c // indirect
	golang.org/x/term v0.0.0-20201210144234-2321bbc49cbf // indirect
	golang.org/x/time v0.0.0-20201208040808-7e3f01d25324 // indirect
	gopkg.in/yaml.v2 v2.4.0
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
	k8s.io/apimachinery v0.20.2
	k8s.io/cli-runtime v0.20.1
	k8s.io/client-go v11.0.1-0.20190805182717-6502b5e7b1b5+incompatible
	k8s.io/klog/v2 v2.5.0 // indirect
	k8s.io/utils v0.0.0-20210111153108-fddb29f9d009 // indirect
)

replace (
	github.com/tektoncd/pipeline => github.com/jenkins-x/pipeline v0.3.2-0.20210118090417-1e821d85abf6
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.20.1
	k8s.io/client-go => k8s.io/client-go v0.20.1
)

go 1.15
