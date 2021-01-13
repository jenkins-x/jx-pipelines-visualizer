package functions

import (
	"github.com/jenkins-x/jx-pipelines-visualizer/internal/version"
)

func AppVersion() string {
	return version.Version
}
