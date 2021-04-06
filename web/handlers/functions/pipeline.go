package functions

import (
	"fmt"
	"strings"
	"text/template"

	jenkinsv1 "github.com/jenkins-x/jx-api/v4/pkg/apis/jenkins.io/v1"
)

func PipelinePullRequestURL(pa *jenkinsv1.PipelineActivity) string {
	if !strings.HasPrefix(pa.Spec.GitBranch, "PR-") {
		return "" // not a PR
	}

	prNumber := strings.TrimPrefix(pa.Spec.GitBranch, "PR-")
	repoURL := strings.TrimSuffix(pa.Spec.GitURL, ".git")
	prURL := fmt.Sprintf("%s/pull/%s", repoURL, prNumber)
	return prURL
}

func PipelinePreviewEnvironmentApplicationURL(pa *jenkinsv1.PipelineActivity) string {
	for _, stage := range pa.Spec.Steps {
		if stage.Preview != nil && stage.Preview.ApplicationURL != "" {
			return stage.Preview.ApplicationURL
		}
	}
	return ""
}

func PipelineTraceURLFunc(pipelineTraceURLTemplate *template.Template) func(pa *jenkinsv1.PipelineActivity) string {
	return func(pa *jenkinsv1.PipelineActivity) string {
		if pipelineTraceURLTemplate == nil {
			return ""
		}

		traceID := pa.Annotations["pipeline.jenkins-x.io/traceID"]
		if traceID == "" {
			return ""
		}

		sb := new(strings.Builder)
		err := pipelineTraceURLTemplate.Execute(sb, map[string]string{
			"TraceID": traceID,
		})
		if err != nil {
			return err.Error()
		}
		return sb.String()
	}
}
