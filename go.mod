module github.com/jenkins-x/jx-pipelines-visualizer

require (
	github.com/Masterminds/sprig/v3 v3.1.0
	github.com/blevesearch/bleve v1.0.9
	github.com/cznic/b v0.0.0-20181122101859-a26611c4d92d // indirect
	github.com/cznic/mathutil v0.0.0-20181122101859-297441e03548 // indirect
	github.com/cznic/strutil v0.0.0-20181122101858-275e90344537 // indirect
	github.com/facebookgo/ensure v0.0.0-20200202191622-63f1cf65ac4c // indirect
	github.com/facebookgo/stack v0.0.0-20160209184415-751773369052 // indirect
	github.com/facebookgo/subset v0.0.0-20200203212716-c811ad88dec4 // indirect
	github.com/gorilla/mux v1.8.0
	github.com/jenkins-x/jx-api/v3 v3.0.0
	github.com/jenkins-x/jx-kube-client/v3 v3.0.0
	github.com/jenkins-x/jx-pipeline v0.0.38
	github.com/jmhodges/levigo v1.0.0 // indirect
	github.com/mitchellh/go-homedir v1.1.0
	github.com/remyoudompheng/bigfft v0.0.0-20200410134404-eec4a21b6bb0 // indirect
	github.com/rickb777/date v1.13.0
	github.com/rs/xid v1.2.1
	github.com/sirupsen/logrus v1.6.0
	github.com/subchord/go-sse v1.0.1
	github.com/tecbot/gorocksdb v0.0.0-20191217155057-f0fad39f321c // indirect
	github.com/unrolled/render v1.0.3
	github.com/urfave/negroni/v2 v2.0.2
	github.com/tektoncd/pipeline v0.16.3
	k8s.io/apimachinery v0.19.2
	k8s.io/client-go v11.0.1-0.20190805182717-6502b5e7b1b5+incompatible
	)

replace (
	github.com/tektoncd/pipeline => github.com/jenkins-x/pipeline v0.0.0-20201002150609-ca0741e5d19a
	k8s.io/client-go => k8s.io/client-go v0.19.2
)

go 1.15