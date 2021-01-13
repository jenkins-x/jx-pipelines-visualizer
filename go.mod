module github.com/jenkins-x/jx-pipelines-visualizer

require (
	github.com/Masterminds/sprig/v3 v3.1.0
	github.com/blevesearch/bleve v1.0.12
	github.com/cznic/b v0.0.0-20181122101859-a26611c4d92d // indirect
	github.com/cznic/mathutil v0.0.0-20181122101859-297441e03548 // indirect
	github.com/cznic/strutil v0.0.0-20181122101858-275e90344537 // indirect
	github.com/facebookgo/ensure v0.0.0-20200202191622-63f1cf65ac4c // indirect
	github.com/facebookgo/stack v0.0.0-20160209184415-751773369052 // indirect
	github.com/facebookgo/subset v0.0.0-20200203212716-c811ad88dec4 // indirect
	github.com/golang/protobuf v1.4.3 // indirect
	github.com/golang/snappy v0.0.2 // indirect
	github.com/googleapis/gnostic v0.4.2 // indirect
	github.com/gorilla/mux v1.8.0
	github.com/jenkins-x/jx-api/v4 v4.0.11
	github.com/jenkins-x/jx-helpers/v3 v3.0.33 // indirect
	github.com/jenkins-x/jx-pipeline v0.0.71
	github.com/jmhodges/levigo v1.0.0 // indirect
	github.com/mattn/go-colorable v0.1.8 // indirect
	github.com/mitchellh/go-homedir v1.1.0
	github.com/remyoudompheng/bigfft v0.0.0-20200410134404-eec4a21b6bb0 // indirect
	github.com/rickb777/date v1.13.0
	github.com/rs/xid v1.2.1
	github.com/sirupsen/logrus v1.7.0
	github.com/subchord/go-sse v1.0.1
	github.com/tecbot/gorocksdb v0.0.0-20191217155057-f0fad39f321c // indirect
	github.com/tektoncd/pipeline v0.16.3
	github.com/tinylib/msgp v1.1.2 // indirect
	github.com/unrolled/render v1.0.3
	github.com/urfave/negroni/v2 v2.0.2
	github.com/willf/bitset v1.1.11 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	gopkg.in/yaml.v2 v2.3.0
	k8s.io/apimachinery v0.19.3
	k8s.io/cli-runtime v0.19.2
	k8s.io/client-go v11.0.1-0.20190805182717-6502b5e7b1b5+incompatible
)

replace (
	github.com/tektoncd/pipeline => github.com/jenkins-x/pipeline v0.0.0-20201002150609-ca0741e5d19a
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.19.2
	k8s.io/client-go => k8s.io/client-go v0.19.2
)

go 1.15
