module github.com/jenkins-x/jx-pipelines-visualizer

require (
	github.com/Masterminds/sprig/v3 v3.1.0
	github.com/blevesearch/bleve v1.0.14
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/snappy v0.0.2 // indirect
	github.com/googleapis/gnostic v0.4.2 // indirect
	github.com/gorilla/mux v1.8.0
	github.com/jenkins-x/jx-api/v4 v4.0.21
	github.com/jenkins-x/jx-helpers/v3 v3.0.62 // indirect
	github.com/jenkins-x/jx-pipeline v0.0.80
	github.com/mitchellh/go-homedir v1.1.0
	github.com/rickb777/date v1.13.0
	github.com/rs/xid v1.2.1
	github.com/sirupsen/logrus v1.7.0
	github.com/subchord/go-sse v1.0.1
	github.com/tektoncd/pipeline v0.16.3
	github.com/tinylib/msgp v1.1.5 // indirect
	github.com/unrolled/render v1.0.3
	github.com/urfave/negroni/v2 v2.0.2
	github.com/willf/bitset v1.1.11 // indirect
	golang.org/x/text v0.3.5 // indirect
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.19.7 // indirect
	k8s.io/apimachinery v0.19.7
	k8s.io/cli-runtime v0.19.2
	k8s.io/client-go v11.0.1-0.20190805182717-6502b5e7b1b5+incompatible
)

replace (
	github.com/tektoncd/pipeline => github.com/jenkins-x/pipeline v0.0.0-20201002150609-ca0741e5d19a
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.19.2
	k8s.io/client-go => k8s.io/client-go v0.19.2
)

go 1.15
