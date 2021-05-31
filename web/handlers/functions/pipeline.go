package functions

import (
	jenkinsv1 "github.com/jenkins-x/jx-api/v4/pkg/apis/jenkins.io/v1"
)

func PipelinePreviewEnvironmentApplicationURL(pa *jenkinsv1.PipelineActivity) string {
	for _, stage := range pa.Spec.Steps {
		if stage.Preview != nil && stage.Preview.ApplicationURL != "" {
			return stage.Preview.ApplicationURL
		}
	}
	return ""
}
